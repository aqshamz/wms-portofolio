# Development-only, additive API seed. Run with the backend already running.
# Never updates/deletes records, grants access, or creates inventory transactions.
# Run seed-study-inventory.ps1 afterwards to add idempotent opening inventory.
[CmdletBinding()]
param(
    [string]$BaseUrl = 'http://localhost:8080',
    [string]$Token = $env:WMS_STUDY_TOKEN
)

$ErrorActionPreference = 'Stop'
$uri = [uri]$BaseUrl
if (-not $uri.IsAbsoluteUri -or -not $uri.IsLoopback -or
    $uri.Scheme -notin @('http', 'https') -or $uri.UserInfo -or
    $uri.AbsolutePath -ne '/' -or $uri.Query -or $uri.Fragment) {
    throw 'Use a loopback development API URL, for example http://localhost:8080.'
}
$BaseUrl = $BaseUrl.TrimEnd('/')
$headers = @{}
$ownedSession = $false
$script:created = 0
$script:reused = 0

function Invoke-Api($Method, $Path, $Body = $null) {
    $params = @{ Uri = "$BaseUrl/api/v1/$Path"; Method = $Method; Headers = $headers; TimeoutSec = 30 }
    if ($null -ne $Body) {
        $params.ContentType = 'application/json'
        $params.Body = $Body | ConvertTo-Json -Depth 12 -Compress
    }
    $response = Invoke-RestMethod @params
    if (-not $response.success) { throw "API failed: $Method $Path" }
    return $response.data
}

function Get-Records($Path) {
    $separator = '?'
    if ($Path.Contains('?')) { $separator = '&' }
    for ($page = 1; $page -le 10000; $page++) {
        $data = Invoke-Api GET "master/$Path${separator}page=$page&page_size=100"
        if ($null -eq $data) { return }
        if ($data.PSObject.Properties.Name -contains 'items') {
            $data.items
            if ($page -ge $data.total_pages) { return }
        } else {
            $data
            return
        }
    }
    throw "Pagination limit reached for $Path"
}

function Assert-Record($Record, [hashtable]$Expected, $Label) {
    if ($null -eq $Record) { throw "Missing record: $Label" }
    if ($Record.PSObject.Properties.Name -contains 'is_active' -and -not $Record.is_active) {
        throw "$Label is inactive. No records were reactivated; review it manually."
    }
    foreach ($key in $Expected.Keys) {
        if ([string]$Record.$key -ne [string]$Expected[$key]) {
            throw "$Label has a different $key. Existing data was left unchanged."
        }
    }
}

# Natural keys are scoped by the collection URL. Relationship assertions prevent
# an existing study code from silently being reused for a different owner/zone.
function Ensure-Record($Path, [hashtable]$Body, [string[]]$Keys = @('code'), [hashtable]$Links = @{}) {
    $matches = @(Get-Records $Path | Where-Object {
        $record = $_
        $same = $true
        foreach ($key in $Keys) { if ([string]$record.$key -ne [string]$Body[$key]) { $same = $false } }
        $same
    })
    $label = "$Path [$($Keys.ForEach({ $Body[$_] }) -join ', ')]"
    if ($matches.Count -gt 1) { throw "Ambiguous natural key: $label" }
    if ($matches.Count -eq 1) {
        Assert-Record $matches[0] $Links $label
        $script:reused++
        return $matches[0]
    }
    # These POSTs can deselect another record. Do not replace user selections
    # when continuing a partially populated study dataset.
    if ($Body.ContainsKey('is_primary') -and $Body.is_primary) {
        if (@(Get-Records $Path | Where-Object is_primary).Count -gt 0) {
            $Body.is_primary = $false
        }
    }
    if ($Body.ContainsKey('is_initial') -and $Body.is_initial) {
        if (@(Get-Records $Path | Where-Object is_initial).Count -gt 0) {
            throw "Another initial status exists in $Path. Review manually."
        }
    }
    Invoke-Api POST "master/$Path" $Body | Out-Null
    # Read back even association endpoints whose POST returns no body.
    $matches = @(Get-Records $Path | Where-Object {
        $record = $_
        $same = $true
        foreach ($key in $Keys) { if ([string]$record.$key -ne [string]$Body[$key]) { $same = $false } }
        $same
    })
    if ($matches.Count -ne 1) { throw "Read-back failed: $label" }
    Assert-Record $matches[0] $Links $label
    $script:created++
    Write-Host "Created $label"
    return $matches[0]
}

function Require-Reference($Path, $Code) {
    $records = @(Get-Records $Path | Where-Object code -EQ $Code)
    if ($records.Count -ne 1) { throw "Expected one seeded reference: $Path/$Code. Start the current backend first." }
    Assert-Record $records[0] @{} "$Path/$Code"
    return $records[0]
}

try {
    if (-not $Token) {
        $credentials = @{}
        Get-Content -LiteralPath (Join-Path $PSScriptRoot '../.env') | ForEach-Object {
            if ($_ -match '^AUTH_BOOTSTRAP_ADMIN_(USERNAME|PASSWORD)=(.*)$') {
                $credentials[$Matches[1]] = $Matches[2].Trim().Trim('"').Trim("'")
            }
        }
        if (-not $credentials.USERNAME -or -not $credentials.PASSWORD) {
            throw 'Set WMS_STUDY_TOKEN or configure bootstrap admin credentials in backend/.env.'
        }
        $login = Invoke-Api POST 'auth/login' @{ identifier = $credentials.USERNAME; password = $credentials.PASSWORD }
        $Token = $login.token
        if (-not $Token) { throw 'Login returned no token.' }
        $ownedSession = $true
        $credentials.Clear()
    }
    $headers.Authorization = "Bearer $Token"

    # Validate shared references before making any master-data changes.
    $ea = Require-Reference 'uoms' 'EA'
    $box = Require-Reference 'uoms' 'BOX'
    $supplierType = Require-Reference 'partner-types' 'SUPPLIER'
    $customerType = Require-Reference 'partner-types' 'CUSTOMER'
    $available = Require-Reference 'inventory-statuses' 'AVAILABLE'
    $fefo = Require-Reference 'picking-sort-methods' 'FEFO'
    Require-Reference 'modules' 'INBOUND' | Out-Null
    $locationTypes = @{}
    foreach ($code in @('RECEIVING', 'STORAGE', 'PICK_FACE', 'STAGING', 'SHIPPING', 'QUARANTINE')) {
        $locationTypes[$code] = Require-Reference 'location-types' $code
    }

    $operator = Ensure-Record 'organizations' @{ code = 'STUDY_OP'; name = 'Study - Warehouse Operator'; timezone_name = 'Asia/Jakarta'; country_code = 'ID' }
    $owner = Ensure-Record 'organizations' @{ code = 'STUDY_OWNER'; name = 'Study - Coffee Client'; timezone_name = 'Asia/Jakarta'; country_code = 'ID' }
    $operatorId = $operator.organization_id
    $ownerId = $owner.organization_id
    $warehouse = Ensure-Record 'warehouses' @{ code = 'STUDY_WH'; name = 'Study - Jakarta Warehouse'; operator_id = $operatorId; timezone_name = 'Asia/Jakarta'; city = 'Jakarta'; country_code = 'ID' } -Links @{ operator_id = $operatorId }
    $warehouseId = $warehouse.warehouse_id
    Ensure-Record "warehouses/$warehouseId/owners" @{ owner_id = $ownerId } @('owner_id') | Out-Null

    $transferWarehouse = Ensure-Record 'warehouses' @{ code = 'STUDY_WH_2'; name = 'Study - Surabaya Warehouse'; operator_id = $operatorId; timezone_name = 'Asia/Jakarta'; city = 'Surabaya'; country_code = 'ID' } -Links @{ operator_id = $operatorId }
    $transferWarehouseId = $transferWarehouse.warehouse_id
    Ensure-Record "warehouses/$transferWarehouseId/owners" @{ owner_id = $ownerId } @('owner_id') | Out-Null
    $transferZone = Ensure-Record "warehouses/$transferWarehouseId/zones" @{ code = 'STUDY_STORAGE'; name = 'Study - Transfer Storage' }
    $transferLinks = @{ zone_id = $transferZone.zone_id; location_type_id = $locationTypes.STORAGE.location_type_id }
    $transferLocation = Ensure-Record "warehouses/$transferWarehouseId/locations" @{ code = 'STUDY_BULK_01'; zone_id = $transferLinks.zone_id; location_type_id = $transferLinks.location_type_id; is_pick_face = $false; pick_sequence = 10 } -Links $transferLinks

    $zones = @{}
    foreach ($code in @('INBOUND', 'STORAGE', 'OUTBOUND')) {
        $zones[$code] = Ensure-Record "warehouses/$warehouseId/zones" @{ code = "STUDY_$code"; name = "Study - $code" }
    }
    $locations = @()
    foreach ($spec in @(
        @('RCV', 'INBOUND', 'RECEIVING'), @('BULK', 'STORAGE', 'STORAGE'),
        @('PICK', 'STORAGE', 'PICK_FACE'), @('STAGE', 'OUTBOUND', 'STAGING'),
        @('SHIP', 'OUTBOUND', 'SHIPPING'), @('QC', 'INBOUND', 'QUARANTINE')
    )) {
        $links = @{ zone_id = $zones[$spec[1]].zone_id; location_type_id = $locationTypes[$spec[2]].location_type_id }
        $body = @{ code = "STUDY_$($spec[0])_01"; zone_id = $links.zone_id; location_type_id = $links.location_type_id; is_pick_face = ($spec[2] -eq 'PICK_FACE'); pick_sequence = 10 }
        $locations += Ensure-Record "warehouses/$warehouseId/locations" $body -Links $links
    }

    $partners = @{}
    foreach ($spec in @(@('SUPPLIER', $supplierType), @('CUSTOMER', $customerType))) {
        $partner = Ensure-Record "business-partners?owner_id=$ownerId" @{ owner_id = $ownerId; code = "STUDY_$($spec[0])"; name = "Study - Coffee $($spec[0])"; country_code = 'ID' } -Links @{ owner_id = $ownerId }
        Ensure-Record "business-partners/$($partner.partner_id)/types" @{ partner_type_id = $spec[1].partner_type_id } @('partner_type_id') | Out-Null
        $partners[$spec[0]] = $partner
    }
    $category = Ensure-Record "item-categories?owner_id=$ownerId" @{ owner_id = $ownerId; code = 'STUDY_COFFEE'; name = 'Study - Coffee' } -Links @{ owner_id = $ownerId }
    $items = @()
    foreach ($spec in @(@('250G', '0.250000', '24'), @('1KG', '1.000000', '6'))) {
        $links = @{ owner_id = $ownerId; category_id = $category.category_id; base_uom_id = $ea.uom_id }
        $item = Ensure-Record "items?owner_id=$ownerId" @{ owner_id = $ownerId; category_id = $category.category_id; base_uom_id = $ea.uom_id; code = "STUDY_COFFEE_$($spec[0])"; name = "Study - Coffee $($spec[0])"; weight = $spec[1]; lot_controlled = $true; serial_controlled = $false; shelf_life_days = 365; minimum_receive_days = 90 } -Links $links
        Ensure-Record "items/$($item.item_id)/uoms" @{ uom_id = $box.uom_id; conversion_to_base = $spec[2]; is_receiving_uom = $true; is_picking_uom = $false } @('uom_id') | Out-Null
        foreach ($unit in @($ea, $box)) {
            Ensure-Record "items/$($item.item_id)/barcodes" @{ uom_id = $unit.uom_id; barcode = "STUDY-COFFEE-$($spec[0])-$($unit.code)"; is_primary = ($unit.code -eq 'EA') } @('barcode') -Links @{ uom_id = $unit.uom_id } | Out-Null
        }
        $items += Invoke-Api GET "master/items/$($item.item_id)"
    }
    $equipmentCategory = Ensure-Record "item-categories?owner_id=$ownerId" @{ owner_id = $ownerId; code = 'STUDY_EQUIPMENT'; name = 'Study - Equipment' } -Links @{ owner_id = $ownerId }
    # A scanner created from the older identity guide may have no category.
    # Reuse it when its owner/UOM/control flags are compatible; fresh datasets
    # place it in STUDY_EQUIPMENT.
    $scannerLinks = @{ owner_id = $ownerId; base_uom_id = $ea.uom_id; serial_controlled = $true; lot_controlled = $false }
    $scanner = Ensure-Record "items?owner_id=$ownerId" @{ owner_id = $ownerId; category_id = $equipmentCategory.category_id; base_uom_id = $ea.uom_id; code = 'STUDY_SCANNER'; name = 'Study - Handheld Scanner'; serial_controlled = $true; lot_controlled = $false } -Links $scannerLinks
    Ensure-Record "items/$($scanner.item_id)/barcodes" @{ uom_id = $ea.uom_id; barcode = 'STUDY-SCANNER-EA'; is_primary = $true } @('barcode') -Links @{ uom_id = $ea.uom_id } | Out-Null
    $items += Invoke-Api GET "master/items/$($scanner.item_id)"

    $scope = @{ owner_id = $ownerId; warehouse_id = $warehouseId }
    $picking = Ensure-Record "picking-strategies?owner_id=$ownerId&warehouse_id=$warehouseId" @{ owner_id = $ownerId; warehouse_id = $warehouseId; code = 'STUDY_FEFO'; name = 'Study - First Expiry First Out' } -Links $scope
    Ensure-Record "picking-strategies/$($picking.picking_strategy_id)/rules" @{ sequence_no = 10; inventory_status_id = $available.inventory_status_id; zone_id = $zones.STORAGE.zone_id; picking_sort_method_id = $fefo.picking_sort_method_id } @('sequence_no') -Links @{ zone_id = $zones.STORAGE.zone_id } | Out-Null
    $putaway = Ensure-Record "putaway-strategies?owner_id=$ownerId&warehouse_id=$warehouseId" @{ owner_id = $ownerId; warehouse_id = $warehouseId; code = 'STUDY_STORAGE'; name = 'Study - Coffee Storage' } -Links $scope
    Ensure-Record "putaway-strategies/$($putaway.putaway_strategy_id)/rules" @{ sequence_no = 10; category_id = $category.category_id; location_type_id = $locationTypes.STORAGE.location_type_id; zone_id = $zones.STORAGE.zone_id; minimum_empty_percent = '25.0000' } @('sequence_no') -Links @{ category_id = $category.category_id; zone_id = $zones.STORAGE.zone_id } | Out-Null

    $document = Ensure-Record 'document-types' @{ code = 'STUDY_RECEIPT'; name = 'Study - Receipt'; module_code = 'INBOUND' } -Links @{ module_code = 'INBOUND' }
    $documentId = $document.document_type_id
    $statuses = @{}
    $order = 0
    foreach ($code in @('DRAFT', 'READY', 'COMPLETED', 'CANCELLED')) {
        $order += 10
        $statuses[$code] = Ensure-Record "document-types/$documentId/statuses" @{ code = $code; name = $code; is_initial = ($code -eq 'DRAFT'); is_final = ($code -in @('COMPLETED', 'CANCELLED')); is_cancelled = ($code -eq 'CANCELLED'); display_order = $order }
    }
    foreach ($edge in @(@('DRAFT', 'READY'), @('DRAFT', 'CANCELLED'), @('READY', 'COMPLETED'), @('READY', 'CANCELLED'))) {
        Ensure-Record "document-types/$documentId/transitions" @{ from_status_id = $statuses[$edge[0]].status_id; to_status_id = $statuses[$edge[1]].status_id } @('from_status_id', 'to_status_id') | Out-Null
    }
    # Number-rule POST replaces the active rule. Never call it on reruns.
    $rules = @(Get-Records "document-types/$documentId/number-rules")
    if ($rules.Count -eq 0) {
        $jakartaNow = [DateTime]::UtcNow.AddHours(7).ToString('yyyy-MM-dd')
        Invoke-Api POST "master/document-types/$documentId/number-rules" @{ prefix = 'STD_RCPT'; separator = '-'; sequence_length = 6; include_partner_code = $true; include_warehouse_code = $true; effective_from = $jakartaNow } | Out-Null
        $script:created++
    } else {
        if (@($rules | Where-Object is_active).Count -ne 1) { throw 'Study document has no unique active number rule. Review manually.' }
        $script:reused++
    }
    $task = Ensure-Record 'task-types' @{ code = 'STUDY_PICK'; name = 'Study - Picking'; description = 'Study master configuration only; no task execution is created.' }
    $numberRules = @(Get-Records "document-types/$documentId/number-rules")
    if (@($numberRules | Where-Object is_active).Count -ne 1) { throw 'Number rule read-back failed.' }
    foreach ($code in @('INBOUND', 'STORAGE', 'OUTBOUND')) {
        $zones[$code] = Get-Records "warehouses/$warehouseId/zones" | Where-Object code -EQ "STUDY_$code"
    }

    # Safe manifest: IDs and study records only, never credentials or tokens.
    [pscustomobject]@{
        created = $script:created; reused = $script:reused
        operator = $operator; owner = $owner; warehouse = $warehouse
        transfer_warehouse = $transferWarehouse; transfer_location = $transferLocation
        zones = $zones; locations = $locations; partners = $partners
        category = $category; categories = @($category, $equipmentCategory); items = $items
        picking_strategy = $picking; putaway_strategy = $putaway
        document_type = $document; document_statuses = $statuses; number_rules = $numberRules; task_type = $task
    }
} finally {
    if ($ownedSession) {
        try { Invoke-Api POST 'auth/logout' | Out-Null }
        catch { Write-Warning 'Could not revoke the seed login session. Log out/revoke it manually.' }
    }
    $Token = $null
    $headers.Clear()
}
