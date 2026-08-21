const fs = require('fs');
const path = require('path');

const workspace = path.resolve(__dirname, '..', '..');
const schemaPath = path.join(workspace, 'database', 'wms_schema.sql');
const reportSetupPath = path.join(workspace, 'database', 'reports', '00_reporting_setup.sql');
const outputDir = __dirname;
const visualizationPath = process.argv[2] || null;

const schema = fs.readFileSync(schemaPath, 'utf8');
const reportSetup = fs.existsSync(reportSetupPath)
  ? fs.readFileSync(reportSetupPath, 'utf8')
  : '';

function splitTopLevel(text) {
  const parts = [];
  let start = 0;
  let depth = 0;
  let single = false;
  let dollarTag = null;
  for (let i = 0; i < text.length; i += 1) {
    if (dollarTag) {
      if (text.startsWith(dollarTag, i)) {
        i += dollarTag.length - 1;
        dollarTag = null;
      }
      continue;
    }
    if (!single && text[i] === '$') {
      const match = text.slice(i).match(/^\$[A-Za-z0-9_]*\$/);
      if (match) {
        dollarTag = match[0];
        i += dollarTag.length - 1;
        continue;
      }
    }
    if (text[i] === "'") {
      if (single && text[i + 1] === "'") {
        i += 1;
      } else {
        single = !single;
      }
      continue;
    }
    if (single) continue;
    if (text[i] === '(') depth += 1;
    if (text[i] === ')') depth -= 1;
    if (text[i] === ',' && depth === 0) {
      parts.push(text.slice(start, i).trim());
      start = i + 1;
    }
  }
  const tail = text.slice(start).trim();
  if (tail) parts.push(tail);
  return parts;
}

function parseList(value) {
  return value.split(',').map((item) => item.trim().replace(/^"|"$/g, ''));
}

function cleanDefault(value) {
  return value
    .replace(/\s+/g, ' ')
    .replace(/\s+(?:CONSTRAINT|CHECK|REFERENCES|UNIQUE|PRIMARY KEY|NOT NULL).*$/i, '')
    .trim();
}

const tables = [];
const tableByName = new Map();
const relationships = [];
const tableRegex = /^CREATE TABLE\s+([a-z_][a-z0-9_]*)\s*\(([\s\S]*?)^\);/gim;
let tableMatch;

while ((tableMatch = tableRegex.exec(schema)) !== null) {
  const name = tableMatch[1];
  const body = tableMatch[2];
  const table = { name, columns: [], primaryKey: [], uniqueKeys: [], domain: '' };
  const entries = splitTopLevel(body);

  for (const rawEntry of entries) {
    const entry = rawEntry.replace(/--.*$/gm, '').trim();
    if (!entry) continue;

    const constraintEntry = entry.replace(/^CONSTRAINT\s+\S+\s+/i, '');
    const pkMatch = constraintEntry.match(/^PRIMARY KEY\s*\(([^)]+)\)/i);
    if (pkMatch) {
      table.primaryKey.push(...parseList(pkMatch[1]));
    }

    const uniqueMatches = [...entry.matchAll(/(?:^|\s)UNIQUE(?:\s+NULLS\s+NOT\s+DISTINCT)?\s*\(([^)]+)\)/gi)];
    for (const uniqueMatch of uniqueMatches) {
      table.uniqueKeys.push(parseList(uniqueMatch[1]));
    }

    const fkMatch = entry.match(/FOREIGN KEY\s*\(([^)]+)\)\s*REFERENCES\s+([a-z_][a-z0-9_]*)\s*\(([^)]+)\)([\s\S]*)/i);
    if (fkMatch) {
      const sourceColumns = parseList(fkMatch[1]);
      relationships.push({
        source: name,
        sourceColumns,
        target: fkMatch[2],
        targetColumns: parseList(fkMatch[3]),
        nullable: sourceColumns.some((columnName) => table.columns.find((column) => column.name === columnName)?.nullable),
        onDelete: (fkMatch[4].match(/ON DELETE\s+(CASCADE|RESTRICT|SET NULL|SET DEFAULT|NO ACTION)/i) || [])[1]?.toUpperCase() || null,
      });
      continue;
    }

    const normalized = entry.replace(/^CONSTRAINT\s+\S+\s+/i, '');
    if (/^(PRIMARY|UNIQUE|FOREIGN|CHECK|EXCLUDE)\b/i.test(normalized)) continue;
    const columnMatch = normalized.match(/^([a-z_][a-z0-9_]*)\s+(.+)$/i);
    if (!columnMatch) continue;

    const columnName = columnMatch[1];
    const remainder = columnMatch[2].trim();
    const constraintIndex = remainder.search(/\s+(?:NOT NULL|NULL\b|DEFAULT\b|PRIMARY KEY|UNIQUE\b|REFERENCES\b|CHECK\b|CONSTRAINT\b|GENERATED\b)/i);
    const type = (constraintIndex === -1 ? remainder : remainder.slice(0, constraintIndex)).trim();
    const constraints = constraintIndex === -1 ? '' : remainder.slice(constraintIndex).trim();
    const inlinePrimary = /\bPRIMARY KEY\b/i.test(constraints);
    const inlineUnique = /\bUNIQUE\b/i.test(constraints);
    const defaultMatch = constraints.match(/\bDEFAULT\s+(.+)$/i);
    const referenceMatch = constraints.match(/\bREFERENCES\s+([a-z_][a-z0-9_]*)\s*\(([^)]+)\)([\s\S]*)/i);
    const column = {
      name: columnName,
      type,
      nullable: !/\bNOT NULL\b/i.test(constraints) && !inlinePrimary,
      primary: inlinePrimary,
      unique: inlineUnique,
      default: defaultMatch ? cleanDefault(defaultMatch[1]) : null,
    };
    table.columns.push(column);
    if (inlinePrimary) table.primaryKey.push(columnName);
    if (inlineUnique) table.uniqueKeys.push([columnName]);
    if (referenceMatch) {
      relationships.push({
        source: name,
        sourceColumns: [columnName],
        target: referenceMatch[1],
        targetColumns: parseList(referenceMatch[2]),
        nullable: column.nullable,
        onDelete: (referenceMatch[3].match(/ON DELETE\s+(CASCADE|RESTRICT|SET NULL|SET DEFAULT|NO ACTION)/i) || [])[1]?.toUpperCase() || null,
      });
    }
  }

  table.primaryKey = [...new Set(table.primaryKey)];
  for (const column of table.columns) {
    if (table.primaryKey.includes(column.name)) {
      column.primary = true;
      column.nullable = false;
    }
  }
  tables.push(table);
  tableByName.set(name, table);
}

const alterFkRegex = /ALTER TABLE\s+([a-z_][a-z0-9_]*)\s+ADD CONSTRAINT\s+\S+\s+FOREIGN KEY\s*\(([^)]+)\)\s+REFERENCES\s+([a-z_][a-z0-9_]*)\s*\(([^)]+)\)([\s\S]*?);/gim;
let alterMatch;
while ((alterMatch = alterFkRegex.exec(schema)) !== null) {
  const sourceTable = tableByName.get(alterMatch[1]);
  const sourceColumns = parseList(alterMatch[2]);
  relationships.push({
    source: alterMatch[1],
    sourceColumns,
    target: alterMatch[3],
    targetColumns: parseList(alterMatch[4]),
    nullable: sourceColumns.some((columnName) => sourceTable?.columns.find((column) => column.name === columnName)?.nullable),
    onDelete: (alterMatch[5].match(/ON DELETE\s+(CASCADE|RESTRICT|SET NULL|SET DEFAULT|NO ACTION)/i) || [])[1]?.toUpperCase() || null,
  });
}

function inRange(name, start, end) {
  const index = tables.findIndex((table) => table.name === name);
  const startIndex = tables.findIndex((table) => table.name === start);
  const endIndex = tables.findIndex((table) => table.name === end);
  return index >= startIndex && index <= endIndex;
}

function domainFor(name) {
  if (inRange(name, 'account_status', 'menu_permission')) return 'security';
  if (inRange(name, 'organization', 'handling_unit_type')) return 'master';
  if (inRange(name, 'document_type', 'document_daily_counter')) return 'configuration';
  if (inRange(name, 'inventory_lot', 'handling_unit')) return 'inventory_identity';
  if (inRange(name, 'purchase_order', 'putaway_task')) return 'inbound';
  if (inRange(name, 'movement_type', 'rework_task')) {
    return ['quarantine_case', 'quarantine_disposition', 'rework_task'].includes(name)
      ? 'inbound'
      : 'stock_control';
  }
  if (inRange(name, 'carrier', 'delivery_return_line')) return 'outbound';
  if (inRange(name, 'internal_move_order', 'stock_count_line')) return 'stock_control';
  if (inRange(name, 'currency', 'billing_payment_allocation')) return 'billing';
  return 'shared';
}

for (const table of tables) table.domain = domainFor(table.name);

const relationshipKeys = new Set();
const dedupedRelationships = [];
for (const relationship of relationships) {
  const key = `${relationship.source}:${relationship.sourceColumns.join(',')}>${relationship.target}:${relationship.targetColumns.join(',')}`;
  if (!relationshipKeys.has(key)) {
    relationshipKeys.add(key);
    dedupedRelationships.push(relationship);
  }
}

for (const relationship of dedupedRelationships) {
  const sourceTable = tableByName.get(relationship.source);
  relationship.nullable = relationship.sourceColumns.some((columnName) =>
    sourceTable?.columns.find((column) => column.name === columnName)?.nullable
  );
}

const views = [...reportSetup.matchAll(/^CREATE OR REPLACE VIEW\s+([a-z_][a-z0-9_]*)\s+AS/gim)]
  .map((match) => match[1]);

const domainOrder = [
  'security', 'master', 'configuration', 'inventory_identity',
  'inbound', 'stock_control', 'outbound', 'billing', 'shared',
];
const domainLabels = {
  security: 'Security & access',
  master: 'Master data',
  configuration: 'Workflow & configuration',
  inventory_identity: 'Inventory identity',
  inbound: 'Inbound',
  stock_control: 'Stock control',
  outbound: 'Outbound',
  billing: 'Billing',
  shared: 'Shared',
};

function escapeDbml(value) {
  return String(value).replace(/\\/g, '\\\\').replace(/'/g, "\\'");
}

function dbml() {
  const lines = [
    '// Generated from database/wms_schema.sql. Do not edit manually.',
    '// Import this file into dbdiagram.io or another DBML-compatible tool.',
    '',
    'Project WMS {',
    "  database_type: 'PostgreSQL'",
    "  Note: 'WMS PostgreSQL 15+ schema; transaction IDs are varchar business identifiers.'",
    '}',
    '',
  ];
  for (const table of tables) {
    lines.push(`Table wms.${table.name} {`);
    for (const column of table.columns) {
      const attributes = [];
      if (table.primaryKey.includes(column.name)) attributes.push('pk');
      if (!column.nullable) attributes.push('not null');
      if (table.uniqueKeys.some((key) => key.length === 1 && key[0] === column.name)) attributes.push('unique');
      if (column.default) attributes.push(`note: 'default ${escapeDbml(column.default)}'`);
      lines.push(`  ${column.name} ${column.type}${attributes.length ? ` [${attributes.join(', ')}]` : ''}`);
    }
    const compositeUniques = table.uniqueKeys.filter((key) => key.length > 1);
    if (compositeUniques.length) {
      lines.push('  indexes {');
      for (const key of compositeUniques) lines.push(`    (${key.join(', ')}) [unique]`);
      lines.push('  }');
    }
    lines.push('}', '');
  }
  for (const relationship of dedupedRelationships) {
    const source = relationship.sourceColumns.length === 1
      ? `wms.${relationship.source}.${relationship.sourceColumns[0]}`
      : `wms.${relationship.source}.(${relationship.sourceColumns.join(', ')})`;
    const target = relationship.targetColumns.length === 1
      ? `wms.${relationship.target}.${relationship.targetColumns[0]}`
      : `wms.${relationship.target}.(${relationship.targetColumns.join(', ')})`;
    lines.push(`Ref: ${source} > ${target}`);
  }
  lines.push('');
  for (const domain of domainOrder) {
    const domainTables = tables.filter((table) => table.domain === domain);
    if (!domainTables.length) continue;
    lines.push(`TableGroup ${domain} {`);
    for (const table of domainTables) lines.push(`  wms.${table.name}`);
    lines.push('}', '');
  }
  return lines.join('\n');
}

function mermaidType(type) {
  return type.replace(/[^A-Za-z0-9_]/g, '_').replace(/_+/g, '_').replace(/_$/g, '');
}

function mermaidFor(selectedTables, title) {
  const selected = new Set(selectedTables.map((table) => table.name));
  const lines = [
    `%% ${title}`,
    'erDiagram',
  ];
  for (const table of selectedTables) {
    lines.push(`  ${table.name} {`);
    for (const column of table.columns) {
      const keys = [];
      if (table.primaryKey.includes(column.name)) keys.push('PK');
      if (dedupedRelationships.some((relationship) =>
        relationship.source === table.name && relationship.sourceColumns.includes(column.name))) keys.push('FK');
      if (table.uniqueKeys.some((key) => key.length === 1 && key[0] === column.name)) keys.push('UK');
      const keyText = keys.length ? ` ${keys.join(',')}` : '';
      const note = column.default ? ` \"default ${column.default.replace(/\"/g, "'")}\"` : '';
      lines.push(`    ${mermaidType(column.type)} ${column.name}${keyText}${note}`);
    }
    lines.push('  }');
  }
  for (const relationship of dedupedRelationships) {
    if (!selected.has(relationship.source) || !selected.has(relationship.target)) continue;
    const parentCardinality = relationship.nullable ? 'o|' : '||';
    lines.push(`  ${relationship.target} ${parentCardinality}--o{ ${relationship.source} : \"${relationship.sourceColumns.join(', ')}\"`);
  }
  return lines.join('\n') + '\n';
}

function dataDictionary() {
  const totalColumns = tables.reduce((sum, table) => sum + table.columns.length, 0);
  const lines = [
    '# WMS data dictionary',
    '',
    `Generated from \`database/wms_schema.sql\`: **${tables.length} tables**, **${totalColumns} columns**, and **${dedupedRelationships.length} foreign-key relationships**.`,
    '',
    'Legend: PK = primary key, FK = foreign key, UK = single-column unique key, NN = not null.',
    '',
  ];
  for (const domain of domainOrder) {
    const domainTables = tables.filter((table) => table.domain === domain);
    if (!domainTables.length) continue;
    lines.push(`## ${domainLabels[domain]}`, '');
    for (const table of domainTables) {
      lines.push(`### \`${table.name}\``, '', '| Column | PostgreSQL type | Null | Key | Default |', '|---|---|:---:|---|---|');
      for (const column of table.columns) {
        const keys = [];
        if (table.primaryKey.includes(column.name)) keys.push('PK');
        if (dedupedRelationships.some((relationship) =>
          relationship.source === table.name && relationship.sourceColumns.includes(column.name))) keys.push('FK');
        if (table.uniqueKeys.some((key) => key.length === 1 && key[0] === column.name)) keys.push('UK');
        lines.push(`| \`${column.name}\` | \`${column.type}\` | ${column.nullable ? 'Y' : 'N'} | ${keys.join(', ')} | ${column.default ? `\`${column.default.replace(/\|/g, '\\|')}\`` : ''} |`);
      }
      const outgoing = dedupedRelationships.filter((relationship) => relationship.source === table.name);
      if (outgoing.length) {
        lines.push('', 'Relationships:');
        for (const relationship of outgoing) {
          lines.push(`- \`${relationship.source}(${relationship.sourceColumns.join(', ')})\` → \`${relationship.target}(${relationship.targetColumns.join(', ')})\`${relationship.nullable ? ' (optional)' : ''}${relationship.onDelete ? `; ON DELETE ${relationship.onDelete}` : ''}`);
        }
      }
      lines.push('');
    }
  }
  if (views.length) {
    lines.push('## Reporting helper views', '');
    for (const view of views) lines.push(`- \`${view}\``);
    lines.push('');
  }
  return lines.join('\n');
}

function relationshipsMarkdown() {
  const lines = [
    '# WMS relationship catalog',
    '',
    '| Child table/column | Parent table/column | Required | Delete action |',
    '|---|---|:---:|---|',
  ];
  for (const relationship of dedupedRelationships.sort((a, b) =>
    `${a.source}:${a.sourceColumns}`.localeCompare(`${b.source}:${b.sourceColumns}`))) {
    lines.push(`| \`${relationship.source}(${relationship.sourceColumns.join(', ')})\` | \`${relationship.target}(${relationship.targetColumns.join(', ')})\` | ${relationship.nullable ? 'No' : 'Yes'} | ${relationship.onDelete || 'RESTRICT/NO ACTION'} |`);
  }
  lines.push('');
  return lines.join('\n');
}

function readme() {
  const counts = Object.fromEntries(domainOrder.map((domain) => [domain, tables.filter((table) => table.domain === domain).length]));
  return `# WMS ERD study package

This package is generated from the PostgreSQL schema and covers all ${tables.length}
tables, their column types, keys, defaults, nullability, and foreign-key
relationships.

## Files

- \`full_erd.dbml\` - complete model for dbdiagram.io or another DBML viewer.
- \`full_erd.mmd\` - complete Mermaid ER diagram source.
- \`domains/*.mmd\` - smaller Mermaid diagrams for focused study.
- \`DATA_DICTIONARY.md\` - every table and column with PostgreSQL data type.
- \`RELATIONSHIPS.md\` - every foreign key and delete behavior.
- \`erd_manifest.json\` - structured source used by the interactive explorer.
- \`generate-erd.js\` - repeatable generator; run after schema changes.

## Suggested study order

1. Security and access (${counts.security} tables)
2. Master data and workflow configuration (${counts.master + counts.configuration} tables)
3. Inventory identity (${counts.inventory_identity} tables)
4. Inbound (${counts.inbound} tables)
5. Stock control (${counts.stock_control} tables)
6. Outbound (${counts.outbound} tables)
7. Billing (${counts.billing} tables)

The full diagram is intentionally large. Use the domain diagrams first, then
consult the relationship catalog for cross-domain foreign keys.

## Regenerate

From the workspace root:

\`node database/erd/generate-erd.js\`
`;
}

function visualizationFragment(manifest) {
  const json = JSON.stringify(manifest).replace(/</g, '\\u003c');
  return `<div id="wms-erd-explorer">
  <h1>WMS ERD explorer</h1>
  <div class="viz-controls">
    <label class="form-label" for="wms-erd-domain">Domain
      <select class="form-select" id="wms-erd-domain"></select>
    </label>
    <label class="form-label" for="wms-erd-search">Find table
      <input class="form-control" id="wms-erd-search" list="wms-erd-tables" placeholder="Type a table name">
      <datalist id="wms-erd-tables"></datalist>
    </label>
  </div>
  <div class="viz-grid wms-erd-stats">
    <div class="card viz-stat"><span class="text-muted">Tables shown</span><span class="viz-stat-value" id="wms-erd-table-count"></span></div>
    <div class="card viz-stat"><span class="text-muted">Columns shown</span><span class="viz-stat-value" id="wms-erd-column-count"></span></div>
    <div class="card viz-stat"><span class="text-muted">Relationships shown</span><span class="viz-stat-value" id="wms-erd-relation-count"></span></div>
  </div>
  <div class="wms-erd-layout">
    <div class="wms-erd-graph"><svg id="wms-erd-svg" role="img" aria-label="Entity relationship graph"></svg></div>
    <section class="card wms-erd-detail" aria-live="polite">
      <h2 id="wms-erd-table-name">Select a table</h2>
      <div class="text-muted" id="wms-erd-table-domain"></div>
      <div class="table-responsive"><table class="table table-sm"><thead><tr><th>Column</th><th>Type</th><th>Key</th><th>Null</th></tr></thead><tbody id="wms-erd-columns"></tbody></table></div>
      <h3>Relationships</h3>
      <ul id="wms-erd-relationships"></ul>
    </section>
  </div>
</div>
<style>
  #wms-erd-explorer { width: 100%; color: var(--foreground); }
  #wms-erd-explorer .wms-erd-stats { margin: 12px 0; }
  #wms-erd-explorer .wms-erd-layout { display: grid; grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr); gap: 16px; align-items: start; }
  #wms-erd-explorer .wms-erd-graph { min-width: 0; }
  #wms-erd-explorer #wms-erd-svg { width: 100%; height: 620px; display: block; }
  #wms-erd-explorer .wms-erd-edge { stroke: var(--border); stroke-width: 1.2; fill: none; }
  #wms-erd-explorer .wms-erd-node rect { fill: color-mix(in srgb, var(--viz-series-1) 14%, var(--background)); stroke: var(--border); stroke-width: 1; }
  #wms-erd-explorer .wms-erd-node text { fill: var(--foreground); font-size: 12px; pointer-events: none; }
  #wms-erd-explorer .wms-erd-node.is-selected rect { fill: var(--primary); stroke: var(--ring); stroke-width: 2; }
  #wms-erd-explorer .wms-erd-node.is-selected text { fill: var(--primary-foreground); }
  #wms-erd-explorer .wms-erd-node { cursor: pointer; }
  #wms-erd-explorer .wms-erd-detail h2 { overflow-wrap: anywhere; }
  #wms-erd-explorer .wms-erd-detail ul { padding-left: 20px; }
  #wms-erd-explorer .wms-erd-detail li { margin: 5px 0; overflow-wrap: anywhere; }
  @media (max-width: 720px) {
    #wms-erd-explorer .wms-erd-layout { grid-template-columns: 1fr; }
    #wms-erd-explorer #wms-erd-svg { height: 520px; }
  }
</style>
<script src="https://cdn.jsdelivr.net/npm/d3@7.9.0/dist/d3.min.js"></script>
<script>
(() => {
  const manifest = ${json};
  const root = document.getElementById('wms-erd-explorer');
  const domainSelect = root.querySelector('#wms-erd-domain');
  const search = root.querySelector('#wms-erd-search');
  const datalist = root.querySelector('#wms-erd-tables');
  const svg = d3.select(root.querySelector('#wms-erd-svg'));
  const labels = manifest.domainLabels;
  const domains = ['all', ...manifest.domainOrder.filter(domain => manifest.tables.some(table => table.domain === domain))];
  domains.forEach(domain => {
    const option = document.createElement('option');
    option.value = domain;
    option.textContent = domain === 'all' ? 'All domains' : labels[domain];
    domainSelect.appendChild(option);
  });
  domainSelect.value = 'inbound';
  manifest.tables.forEach(table => {
    const option = document.createElement('option');
    option.value = table.name;
    datalist.appendChild(option);
  });

  let selectedName = null;
  let currentTables = [];
  let currentRelationships = [];

  function keyText(table, column) {
    const keys = [];
    if (table.primaryKey.includes(column.name)) keys.push('PK');
    if (manifest.relationships.some(rel => rel.source === table.name && rel.sourceColumns.includes(column.name))) keys.push('FK');
    if (table.uniqueKeys.some(key => key.length === 1 && key[0] === column.name)) keys.push('UK');
    return keys.join(', ');
  }

  function showTable(name) {
    const table = manifest.tables.find(candidate => candidate.name === name);
    if (!table) return;
    selectedName = name;
    root.querySelector('#wms-erd-table-name').textContent = table.name;
    root.querySelector('#wms-erd-table-domain').textContent = labels[table.domain];
    const body = root.querySelector('#wms-erd-columns');
    body.replaceChildren(...table.columns.map(column => {
      const row = document.createElement('tr');
      const values = [column.name, column.type, keyText(table, column), column.nullable ? 'Yes' : 'No'];
      values.forEach((value, index) => {
        const cell = document.createElement('td');
        cell.textContent = value;
        if (index === 0) {
          const code = document.createElement('code');
          code.textContent = value;
          cell.replaceChildren(code);
        }
        row.appendChild(cell);
      });
      return row;
    }));
    const relList = root.querySelector('#wms-erd-relationships');
    const related = manifest.relationships.filter(rel => rel.source === name || rel.target === name);
    relList.replaceChildren(...related.map(rel => {
      const item = document.createElement('li');
      item.textContent = rel.source === name
        ? 'References ' + rel.target + ' via ' + rel.sourceColumns.join(', ')
        : 'Referenced by ' + rel.source + ' via ' + rel.sourceColumns.join(', ');
      return item;
    }));
    svg.selectAll('.wms-erd-node').classed('is-selected', node => node.id === name);
  }

  function draw() {
    const domain = domainSelect.value;
    currentTables = manifest.tables.filter(table => domain === 'all' || table.domain === domain);
    const names = new Set(currentTables.map(table => table.name));
    currentRelationships = manifest.relationships.filter(rel => names.has(rel.source) && names.has(rel.target));
    root.querySelector('#wms-erd-table-count').textContent = currentTables.length;
    root.querySelector('#wms-erd-column-count').textContent = currentTables.reduce((sum, table) => sum + table.columns.length, 0);
    root.querySelector('#wms-erd-relation-count').textContent = currentRelationships.length;

    const element = root.querySelector('#wms-erd-svg');
    const width = Math.max(320, element.getBoundingClientRect().width || 640);
    const height = width < 520 ? 520 : 620;
    svg.attr('viewBox', '0 0 ' + width + ' ' + height);
    svg.selectAll('*').remove();

    const nodes = currentTables.map(table => ({ id: table.name, table, width: Math.max(88, table.name.length * 7 + 20), height: 30 }));
    const links = currentRelationships.map(rel => ({ source: rel.source, target: rel.target, rel }));
    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(links).id(node => node.id).distance(domain === 'all' ? 70 : 95).strength(0.7))
      .force('charge', d3.forceManyBody().strength(domain === 'all' ? -170 : -280))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(node => Math.max(node.width / 2, 46) + 8))
      .stop();
    for (let i = 0; i < (domain === 'all' ? 360 : 240); i += 1) simulation.tick();
    nodes.forEach(node => {
      node.x = Math.max(node.width / 2 + 8, Math.min(width - node.width / 2 - 8, node.x));
      node.y = Math.max(node.height / 2 + 8, Math.min(height - node.height / 2 - 8, node.y));
    });

    svg.append('g').selectAll('line').data(links).join('line')
      .attr('class', 'wms-erd-edge')
      .attr('x1', link => link.source.x).attr('y1', link => link.source.y)
      .attr('x2', link => link.target.x).attr('y2', link => link.target.y);

    const node = svg.append('g').selectAll('g').data(nodes).join('g')
      .attr('class', item => 'wms-erd-node' + (item.id === selectedName ? ' is-selected' : ''))
      .attr('transform', item => 'translate(' + item.x + ',' + item.y + ')')
      .attr('data-tooltip', item => item.id + ': ' + item.table.columns.length + ' columns')
      .on('click', (_, item) => showTable(item.id));
    node.append('rect')
      .attr('x', item => -item.width / 2).attr('y', item => -item.height / 2)
      .attr('width', item => item.width).attr('height', item => item.height).attr('rx', 4);
    node.append('text').attr('text-anchor', 'middle').attr('dy', '0.35em').text(item => item.id);

    if (!selectedName || !names.has(selectedName)) {
      showTable(currentTables[0]?.name);
    } else {
      showTable(selectedName);
    }
  }

  domainSelect.addEventListener('change', draw);
  search.addEventListener('change', () => {
    const table = manifest.tables.find(candidate => candidate.name === search.value.trim());
    if (!table) return;
    domainSelect.value = table.domain;
    selectedName = table.name;
    draw();
  });
  new ResizeObserver(draw).observe(root.querySelector('.wms-erd-graph'));
  draw();
})();
</script>`;
}

const manifest = {
  generatedFrom: 'database/wms_schema.sql',
  tables,
  relationships: dedupedRelationships,
  views,
  domainOrder,
  domainLabels,
};

fs.mkdirSync(outputDir, { recursive: true });
fs.mkdirSync(path.join(outputDir, 'domains'), { recursive: true });
fs.writeFileSync(path.join(outputDir, 'full_erd.dbml'), dbml());
fs.writeFileSync(path.join(outputDir, 'full_erd.mmd'), mermaidFor(tables, 'Complete WMS ERD'));
fs.writeFileSync(path.join(outputDir, 'DATA_DICTIONARY.md'), dataDictionary());
fs.writeFileSync(path.join(outputDir, 'RELATIONSHIPS.md'), relationshipsMarkdown());
fs.writeFileSync(path.join(outputDir, 'erd_manifest.json'), JSON.stringify(manifest, null, 2));
fs.writeFileSync(path.join(outputDir, 'README.md'), readme());
for (const domain of domainOrder) {
  const domainTables = tables.filter((table) => table.domain === domain);
  if (domainTables.length) {
    fs.writeFileSync(
      path.join(outputDir, 'domains', `${domain}.mmd`),
      mermaidFor(domainTables, `${domainLabels[domain]} ERD`)
    );
  }
}
if (visualizationPath) {
  fs.mkdirSync(path.dirname(visualizationPath), { recursive: true });
  fs.writeFileSync(visualizationPath, visualizationFragment(manifest));
}

const totalColumns = tables.reduce((sum, table) => sum + table.columns.length, 0);
process.stdout.write(JSON.stringify({
  tables: tables.length,
  columns: totalColumns,
  relationships: dedupedRelationships.length,
  views: views.length,
  domains: Object.fromEntries(domainOrder.map((domain) => [domain, tables.filter((table) => table.domain === domain).length])),
  visualizationPath,
}, null, 2));
