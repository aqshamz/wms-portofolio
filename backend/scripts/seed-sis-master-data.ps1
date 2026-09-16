# Development-only, additive setup for the SIS owner and Bandung warehouse.
# Existing records are reused. Conflicting natural keys stop the script instead
# of silently overwriting operational master data.
[CmdletBinding()]
param(
    [string]$BaseUrl = 'http://localhost:8080',
    [string]$Token = $env:WMS_SIS_TOKEN
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
$script:updated = 0
$script:reused = 0
$script:newCredentials = @()

function Invoke-Api($Method, $Path, $Body = $null) {
    $params = @{
        Uri = "$BaseUrl/api/v1/$Path"
        Method = $Method
        Headers = $headers
        TimeoutSec = 30
    }
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
        $data = Invoke-Api GET "$Path${separator}page=$page&page_size=100"
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
        throw "$Label is inactive. Review it manually."
    }
    foreach ($key in $Expected.Keys) {
        if ([string]$Record.$key -ne [string]$Expected[$key]) {
            throw "$Label has a different $key. Existing data was left unchanged."
        }
    }
}

function Ensure-Record(
    [string]$ListPath,
    [string]$CreatePath,
    [hashtable]$Body,
    [string[]]$Keys = @('code'),
    [hashtable]$Links = @{}
) {
    $matches = @(Get-Records $ListPath | Where-Object {
        $record = $_
        $same = $true
        foreach ($key in $Keys) {
            if ([string]$record.$key -ne [string]$Body[$key]) { $same = $false }
        }
        $same
    })
    $label = "$ListPath [$($Keys.ForEach({ $Body[$_] }) -join ', ')]"
    if ($matches.Count -gt 1) { throw "Ambiguous natural key: $label" }
    if ($matches.Count -eq 1) {
        Assert-Record $matches[0] $Links $label
        $script:reused++
        return $matches[0]
    }

    Invoke-Api POST $CreatePath $Body | Out-Null
    $matches = @(Get-Records $ListPath | Where-Object {
        $record = $_
        $same = $true
        foreach ($key in $Keys) {
            if ([string]$record.$key -ne [string]$Body[$key]) { $same = $false }
        }
        $same
    })
    if ($matches.Count -ne 1) { throw "Read-back failed: $label" }
    Assert-Record $matches[0] $Links $label
    $script:created++
    Write-Host "Created $label"
    return $matches[0]
}

function Require-One($Path, $Code) {
    $matches = @(Get-Records $Path | Where-Object code -EQ $Code)
    if ($matches.Count -ne 1) { throw "Expected one active record: $Path/$Code" }
    Assert-Record $matches[0] @{} "$Path/$Code"
    return $matches[0]
}

function Ensure-RolePermissions($Role, [string[]]$PermissionCodes, $PermissionsByCode) {
    $detail = Invoke-Api GET "security/roles/$($Role.role_id)"
    $current = @($detail.permissions.code | Sort-Object)
    $wanted = @($PermissionCodes | Sort-Object)
    if (($current -join '|') -eq ($wanted -join '|')) {
        $script:reused++
        return
    }
    $ids = @($PermissionCodes | ForEach-Object {
        if (-not $PermissionsByCode.ContainsKey($_)) { throw "Missing permission: $_" }
        $PermissionsByCode[$_].permission_id
    })
    Invoke-Api PUT "security/roles/$($Role.role_id)/permissions" @{
        permission_ids = $ids
        expected_version = $detail.version_no
    } | Out-Null
    $script:updated++
    Write-Host "Updated role permissions [$($Role.code)]"
}

function Ensure-Role($Code, $Name, $Description, [string[]]$PermissionCodes, $PermissionsByCode) {
    $matches = @(Get-Records 'security/roles' | Where-Object code -EQ $Code)
    if ($matches.Count -gt 1) { throw "Ambiguous role code: $Code" }
    if ($matches.Count -eq 0) {
        $ids = @($PermissionCodes | ForEach-Object {
            if (-not $PermissionsByCode.ContainsKey($_)) { throw "Missing permission: $_" }
            $PermissionsByCode[$_].permission_id
        })
        Invoke-Api POST 'security/roles' @{
            code = $Code
            name = $Name
            description = $Description
            permission_ids = $ids
        } | Out-Null
        $matches = @(Get-Records 'security/roles' | Where-Object code -EQ $Code)
        if ($matches.Count -ne 1) { throw "Role read-back failed: $Code" }
        $script:created++
        Write-Host "Created role [$Code]"
    }
    $role = $matches[0]
    Ensure-RolePermissions $role $PermissionCodes $PermissionsByCode
    return @(Get-Records 'security/roles' | Where-Object code -EQ $Code)[0]
}

function New-InitialPassword {
    $bytes = New-Object byte[] 18
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $suffix = [Convert]::ToBase64String($bytes).Replace('/', 'x').Replace('+', 'Y').TrimEnd('=')
    return "Wm9!$suffix"
}

function Ensure-OperationalAccount($Username, $DisplayName, $Role, $Owner, $Warehouse, $ActiveStatus) {
    $matches = @(Get-Records 'security/accounts' | Where-Object username -EQ $Username)
    if ($matches.Count -gt 1) { throw "Ambiguous account username: $Username" }
    if ($matches.Count -eq 0) {
        $password = New-InitialPassword
        Invoke-Api POST 'security/accounts' @{
            username = $Username
            display_name = $DisplayName
            password = $password
            preferred_timezone = 'Asia/Jakarta'
        } | Out-Null
        $matches = @(Get-Records 'security/accounts' | Where-Object username -EQ $Username)
        if ($matches.Count -ne 1) { throw "Account read-back failed: $Username" }
        $script:newCredentials += [pscustomobject]@{
            username = $Username
            temporary_password = $password
        }
        $script:created++
        Write-Host "Created account [$Username]"
    }
    Ensure-AccountSetup $matches[0] $Role $Owner $Warehouse $ActiveStatus
    return $matches[0]
}

function Ensure-AccountSetup($Account, $Role, $Owner, $Warehouse, $ActiveStatus) {
    $detail = Invoke-Api GET "security/accounts/$($Account.account_id)"
    if (@($detail.roles | Where-Object role_id -EQ $Role.role_id).Count -eq 0) {
        Invoke-Api POST "security/accounts/$($Account.account_id)/roles" @{ role_id = $Role.role_id } | Out-Null
        $script:created++
        Write-Host "Assigned role [$($Role.code)] to [$($Account.username)]"
    } else { $script:reused++ }

    $detail = Invoke-Api GET "security/accounts/$($Account.account_id)"
    if (@($detail.owner_access | Where-Object owner_id -EQ $Owner.organization_id).Count -eq 0) {
        Invoke-Api POST "security/accounts/$($Account.account_id)/owners" @{ owner_id = $Owner.organization_id } | Out-Null
        $script:created++
        Write-Host "Granted owner [$($Owner.code)] to [$($Account.username)]"
    } else { $script:reused++ }

    $detail = Invoke-Api GET "security/accounts/$($Account.account_id)"
    if (@($detail.warehouse_access | Where-Object warehouse_id -EQ $Warehouse.warehouse_id).Count -eq 0) {
        Invoke-Api POST "security/accounts/$($Account.account_id)/warehouses" @{ warehouse_id = $Warehouse.warehouse_id } | Out-Null
        $script:created++
        Write-Host "Granted warehouse [$($Warehouse.code)] to [$($Account.username)]"
    } else { $script:reused++ }

    $detail = Invoke-Api GET "security/accounts/$($Account.account_id)"
    if ($detail.status.code -ne 'ACTIVE') {
        Invoke-Api PATCH "security/accounts/$($Account.account_id)/status" @{
            account_status_id = $ActiveStatus.account_status_id
            expected_version = $detail.version_no
        } | Out-Null
        $script:updated++
        Write-Host "Activated account [$($Account.username)]"
    } else { $script:reused++ }
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
            throw 'Set WMS_SIS_TOKEN or configure bootstrap admin credentials in backend/.env.'
        }
        $login = Invoke-Api POST 'auth/login' @{
            identifier = $credentials.USERNAME
            password = $credentials.PASSWORD
        }
        $Token = $login.token
        if (-not $Token) { throw 'Login returned no token.' }
        $ownedSession = $true
        $credentials.Clear()
    }
    $headers.Authorization = "Bearer $Token"

    $owner = Require-One 'master/organizations' 'SIS'
    $warehouse = Require-One 'master/warehouses' 'BDG_WH'
    $ownerId = $owner.organization_id
    $warehouseId = $warehouse.warehouse_id

    # A warehouse can serve owners other than its operator. This link is also
    # required by owner-and-warehouse-scoped picking/putaway strategies.
    Ensure-Record "master/warehouses/$warehouseId/owners" "master/warehouses/$warehouseId/owners" `
        @{ owner_id = $ownerId } @('owner_id') | Out-Null

    $locationTypes = @{}
    foreach ($code in @('RECEIVING', 'STORAGE', 'PICK_FACE', 'STAGING', 'CHECKING', 'PACKING', 'SHIPPING', 'QUARANTINE')) {
        $locationTypes[$code] = Require-One 'master/location-types' $code
    }
    $zones = @{}
    foreach ($code in @('BDGABT', 'BDGCHL', 'BDGFZR', 'BDGPCK', 'BDGQUA', 'BDGSTG12')) {
        $zones[$code] = Require-One "master/warehouses/$warehouseId/zones" $code
    }

    # Baseline physical layout. Capacities are intentionally omitted because
    # rack engineering values must not be guessed.
    $locationSpecs = @()
    foreach ($i in 1..8) {
        $locationSpecs += @{ code = ('BDG-AMB-BULK-{0:D2}' -f $i); zone = 'BDGABT'; type = 'STORAGE'; pick = $false; sequence = 100 + $i }
    }
    foreach ($i in 1..4) {
        $locationSpecs += @{ code = ('BDG-CHL-BULK-{0:D2}' -f $i); zone = 'BDGCHL'; type = 'STORAGE'; pick = $false; sequence = 200 + $i }
        $locationSpecs += @{ code = ('BDG-FZR-BULK-{0:D2}' -f $i); zone = 'BDGFZR'; type = 'STORAGE'; pick = $false; sequence = 300 + $i }
        $locationSpecs += @{ code = ('BDG-QUA-{0:D2}' -f $i); zone = 'BDGQUA'; type = 'QUARANTINE'; pick = $false; sequence = 500 + $i }
    }
    foreach ($i in 1..8) {
        $locationSpecs += @{ code = ('BDG-PICK-{0:D2}' -f $i); zone = 'BDGPCK'; type = 'PICK_FACE'; pick = $true; sequence = 400 + $i }
    }
    foreach ($spec in @(
        @('BDG-RCV-DOCK-01', 'RECEIVING', 10), @('BDG-RCV-DOCK-02', 'RECEIVING', 20),
        @('BDG-STAGE-IN-01', 'STAGING', 30), @('BDG-STAGE-OUT-01', 'STAGING', 40),
        @('BDG-CHECK-01', 'CHECKING', 50),
        @('BDG-PACK-01', 'PACKING', 60), @('BDG-PACK-02', 'PACKING', 70),
        @('BDG-SHIP-01', 'SHIPPING', 80), @('BDG-SHIP-02', 'SHIPPING', 90)
    )) {
        $locationSpecs += @{ code = $spec[0]; zone = 'BDGSTG12'; type = $spec[1]; pick = $false; sequence = $spec[2] }
    }
    foreach ($spec in $locationSpecs) {
        $links = @{
            zone_id = $zones[$spec.zone].zone_id
            location_type_id = $locationTypes[$spec.type].location_type_id
        }
        $body = @{
            zone_id = $links.zone_id
            location_type_id = $links.location_type_id
            code = $spec.code
            barcode = "LOC-$($spec.code)"
            pick_sequence = $spec.sequence
            is_pick_face = $spec.pick
        }
        Ensure-Record "master/warehouses/$warehouseId/locations" "master/warehouses/$warehouseId/locations" `
            $body @('code') $links | Out-Null
    }

    # Complete a practical two-level SIS catalog taxonomy without inventing SKUs.
    $categories = @{}
    foreach ($spec in @(
        @('ENERGY', 'ENERGY'), @('BETA', 'BETA'), @('SUPLEMENTS', 'SUPLEMENTS'),
        @('HYDRATION', 'HYDRATION'), @('RECOVERY', 'RECOVERY'),
        @('PROTEIN', 'PROTEIN'), @('ACCESSORIES', 'ACCESSORIES')
    )) {
        $categories[$spec[0]] = Ensure-Record "master/item-categories?owner_id=$ownerId" 'master/item-categories' `
            @{ owner_id = $ownerId; code = $spec[0]; name = $spec[1] } @('code') @{ owner_id = $ownerId }
    }
    foreach ($spec in @(
        @('ENERGYCHEWS', 'ENERGY CHEWS', 'ENERGY'),
        @('BETAGELS', 'BETA GELS', 'BETA'), @('BETAPOWDER', 'BETA POWDER', 'BETA'),
        @('ELECTROLYTES', 'ELECTROLYTES', 'HYDRATION'),
        @('HYDRATIONPOWDER', 'HYDRATION POWDER', 'HYDRATION'),
        @('HYDRATIONTABLETS', 'HYDRATION TABLETS', 'HYDRATION'),
        @('RECOVERYPOWDER', 'RECOVERY POWDER', 'RECOVERY'),
        @('RECOVERYBARS', 'RECOVERY BARS', 'RECOVERY'),
        @('PROTEINPOWDER', 'PROTEIN POWDER', 'PROTEIN'),
        @('PROTEINBARS', 'PROTEIN BARS', 'PROTEIN'),
        @('BOTTLES', 'BOTTLES', 'ACCESSORIES'),
        @('SHAKERS', 'SHAKERS', 'ACCESSORIES'),
        @('APPAREL', 'APPAREL', 'ACCESSORIES')
    )) {
        $parent = $categories[$spec[2]]
        Ensure-Record "master/item-categories?owner_id=$ownerId" 'master/item-categories' `
            @{ owner_id = $ownerId; code = $spec[0]; name = $spec[1]; parent_category_id = $parent.category_id } `
            @('code') @{ owner_id = $ownerId; parent_category_id = $parent.category_id } | Out-Null
    }

    # Keep products on leaf categories. These are representative SIS catalogue
    # records for operation-flow development, not a claim that every regional
    # commercial variant/flavour is stocked.
    $allCategories = @{}
    foreach ($category in @(Get-Records "master/item-categories?owner_id=$ownerId")) {
        $allCategories[$category.code] = $category
    }
    $ea = Require-One 'master/uoms' 'EA'
    $itemSpecs = @(
        @{ code = 'SISGOENERGYPOWDER'; name = 'SiS GO Energy Powder 1.6kg'; category = 'ENERGYPOWDER'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISGOENERGYCHEWS'; name = 'SiS GO Energy Chews'; category = 'ENERGYCHEWS'; pack = 'BAR'; lot = $true },
        @{ code = 'SISBETAFUELCHEWS'; name = 'SiS BETA Fuel Chews'; category = 'BETAFUEL'; pack = 'BAR'; lot = $true },
        @{ code = 'SISBETARECOVERY'; name = 'SiS BETA Recovery'; category = 'BETARECOVERY'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISBETAFUELGEL'; name = 'SiS BETA Fuel Energy Gel'; category = 'BETAGELS'; pack = 'GEL'; lot = $true },
        @{ code = 'SISBETAFUELPOWDER'; name = 'SiS BETA Fuel Energy Drink Powder'; category = 'BETAPOWDER'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISBIOTICCOMPLEX'; name = 'SiS Biotic Complex'; category = 'BIOTIC'; pack = 'SUPPLEMENT'; lot = $true },
        @{ code = 'SISADVMULTIVITAMIN'; name = 'SiS Advanced Multivitamin'; category = 'MULTIVITAMINS'; pack = 'SUPPLEMENT'; lot = $true },
        @{ code = 'SISBETAELECTROGEL'; name = 'SiS BETA Fuel + Electrolyte Gel'; category = 'ELECTROLYTES'; pack = 'GEL'; lot = $true },
        @{ code = 'SISHYDROPOWDER'; name = 'SiS GO Electrolyte Powder 500g'; category = 'HYDRATIONPOWDER'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISHYDROTABS'; name = 'SiS HYDRO Electrolyte Tablets'; category = 'HYDRATIONTABLETS'; pack = 'TABLET'; lot = $true },
        @{ code = 'SISREGORAPID'; name = 'SiS REGO Rapid Recovery'; category = 'RECOVERYPOWDER'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISPROTEIN20BAR'; name = 'SiS Protein20 Bar'; category = 'RECOVERYBARS'; pack = 'BAR'; lot = $true },
        @{ code = 'SISREGOWHEY'; name = 'SiS REGO Whey Protein'; category = 'PROTEINPOWDER'; pack = 'POWDER'; lot = $true },
        @{ code = 'SISPROTEINBAR64'; name = 'SiS Protein Bar 64g'; category = 'PROTEINBARS'; pack = 'BAR'; lot = $true },
        @{ code = 'SISBOTTLE800'; name = 'SiS Sports Bottle 800ml'; category = 'BOTTLES'; pack = 'ACCESSORY'; lot = $false },
        @{ code = 'SISSHAKER700'; name = 'SiS Protein Shaker 700ml'; category = 'SHAKERS'; pack = 'ACCESSORY'; lot = $false },
        @{ code = 'SISTSHIRTBLACK'; name = 'SiS Performance T-Shirt Black'; category = 'APPAREL'; pack = 'APPAREL'; lot = $false }
    )
    $itemPackByCode = @{
        SISOATBARS = 'BAR'
        SISGELMAG = 'GEL'
        SISGELAPP = 'GEL'
        SISFRUITBARS = 'BAR'
    }
    foreach ($spec in $itemSpecs) {
        if (-not $allCategories.ContainsKey($spec.category)) {
            throw "Missing leaf category: $($spec.category)"
        }
        $category = $allCategories[$spec.category]
        Ensure-Record "master/items?owner_id=$ownerId" 'master/items' @{
            owner_id = $ownerId
            category_id = $category.category_id
            code = $spec.code
            name = $spec.name
            base_uom_id = $ea.uom_id
            lot_controlled = $spec.lot
            serial_controlled = $false
        } @('code') @{
            owner_id = $ownerId
            category_id = $category.category_id
            base_uom_id = $ea.uom_id
            lot_controlled = $spec.lot
            serial_controlled = $false
        } | Out-Null
        $itemPackByCode[$spec.code] = $spec.pack
    }

    # Packaging is deliberately standardized for the development catalogue:
    # gel 30/120, bar 12/48, powder or supplement 6/24, tablet or accessory
    # 12/48, and apparel 10/40. Replace these before production if supplier
    # case packs differ. Reruns stop rather than overwriting a different value.
    $uoms = @{
        BOX = Require-One 'master/uoms' 'BOX'
        CTN = Require-One 'master/uoms' 'CTN'
    }
    $packProfiles = @{
        GEL = @{ BOX = '30'; CTN = '120' }
        BAR = @{ BOX = '12'; CTN = '48' }
        POWDER = @{ BOX = '6'; CTN = '24' }
        SUPPLEMENT = @{ BOX = '6'; CTN = '24' }
        TABLET = @{ BOX = '12'; CTN = '48' }
        ACCESSORY = @{ BOX = '12'; CTN = '48' }
        APPAREL = @{ BOX = '10'; CTN = '40' }
    }
    $items = @(Get-Records "master/items?owner_id=$ownerId")
    foreach ($item in $items) {
        if (-not $itemPackByCode.ContainsKey($item.code)) { continue }
        $pack = $packProfiles[$itemPackByCode[$item.code]]
        foreach ($uomCode in @('BOX', 'CTN')) {
            $body = @{
                uom_id = $uoms[$uomCode].uom_id
                conversion_to_base = $pack[$uomCode]
                is_receiving_uom = $true
                is_picking_uom = ($uomCode -eq 'BOX')
            }
            Ensure-Record "master/items/$($item.item_id)/uoms" "master/items/$($item.item_id)/uoms" `
                $body @('uom_id') @{ conversion_to_base = "$($pack[$uomCode]).000000" } | Out-Null
        }
    }

    # Business-partner "roles" are partner types in this data model. Keep the
    # full SIS mapping additive; partner login accounts are intentionally not
    # created because the authorization model has no partner-level scope.
    $partnerTypes = @{}
    foreach ($code in @('FACTORY', 'SUPPLIER', 'STORE', 'CUSTOMER', 'CARRIER')) {
        $partnerTypes[$code] = Require-One 'master/partner-types' $code
    }
    $partnerTypeMap = @{
        SIS_FACT_CIKARANG = @('FACTORY', 'SUPPLIER')
        SIS_FACT_CIMAHI = @('FACTORY', 'SUPPLIER')
        SISBINTARO = @('STORE')
        SISJKT = @('STORE')
        RUNDEPT = @('CUSTOMER')
        RANKSPORT = @('CUSTOMER')
        JNE = @('CARRIER')
        SICEPAT = @('CARRIER')
    }
    $partners = @(Get-Records "master/business-partners?owner_id=$ownerId")
    foreach ($partnerCode in $partnerTypeMap.Keys) {
        $matches = @($partners | Where-Object code -EQ $partnerCode)
        if ($matches.Count -ne 1) { throw "Expected existing SIS business partner: $partnerCode" }
        $partner = $matches[0]
        foreach ($typeCode in $partnerTypeMap[$partnerCode]) {
            $type = $partnerTypes[$typeCode]
            Ensure-Record "master/business-partners/$($partner.partner_id)/types" `
                "master/business-partners/$($partner.partner_id)/types" @{
                    partner_type_id = $type.partner_type_id
                } @('partner_type_id') | Out-Null
        }
    }

    $available = Require-One 'master/inventory-statuses' 'AVAILABLE'
    $sortMethods = @{}
    foreach ($code in @('FEFO', 'FIFO', 'LOCATION', 'LOT')) {
        $sortMethods[$code] = Require-One 'master/picking-sort-methods' $code
        $strategy = Ensure-Record "master/picking-strategies?owner_id=$ownerId&warehouse_id=$warehouseId" `
            'master/picking-strategies' @{
                owner_id = $ownerId
                warehouse_id = $warehouseId
                code = "SIS_BDG_$code"
                name = "SIS Bandung - $code"
                description = "Warehouse-scoped $code picking profile for SIS."
            } @('code') @{ owner_id = $ownerId; warehouse_id = $warehouseId }
        Ensure-Record "master/picking-strategies/$($strategy.picking_strategy_id)/rules" `
            "master/picking-strategies/$($strategy.picking_strategy_id)/rules" @{
                sequence_no = 10
                inventory_status_id = $available.inventory_status_id
                picking_sort_method_id = $sortMethods[$code].picking_sort_method_id
            } @('sequence_no') @{ picking_sort_method_id = $sortMethods[$code].picking_sort_method_id } | Out-Null
    }

    $putaway = Ensure-Record "master/putaway-strategies?owner_id=$ownerId&warehouse_id=$warehouseId" `
        'master/putaway-strategies' @{
            owner_id = $ownerId
            warehouse_id = $warehouseId
            code = 'SIS_BDG_STANDARD'
            name = 'SIS Bandung - Standard Putaway'
            description = 'Ambient-first category rules with an ambient fallback.'
        } @('code') @{ owner_id = $ownerId; warehouse_id = $warehouseId }
    $sequence = 0
    foreach ($categoryCode in @('ENERGY', 'BETA', 'SUPLEMENTS', 'HYDRATION', 'RECOVERY', 'PROTEIN', 'ACCESSORIES')) {
        $sequence += 10
        Ensure-Record "master/putaway-strategies/$($putaway.putaway_strategy_id)/rules" `
            "master/putaway-strategies/$($putaway.putaway_strategy_id)/rules" @{
                sequence_no = $sequence
                category_id = $categories[$categoryCode].category_id
                location_type_id = $locationTypes.STORAGE.location_type_id
                zone_id = $zones.BDGABT.zone_id
                minimum_empty_percent = '20.0000'
            } @('sequence_no') @{ category_id = $categories[$categoryCode].category_id; zone_id = $zones.BDGABT.zone_id } | Out-Null
    }
    Ensure-Record "master/putaway-strategies/$($putaway.putaway_strategy_id)/rules" `
        "master/putaway-strategies/$($putaway.putaway_strategy_id)/rules" @{
            sequence_no = 90
            location_type_id = $locationTypes.STORAGE.location_type_id
            zone_id = $zones.BDGABT.zone_id
            minimum_empty_percent = '20.0000'
        } @('sequence_no') @{ zone_id = $zones.BDGABT.zone_id } | Out-Null

    # Normalize least-privilege operational roles, then scope and activate the
    # internal SIS accounts. Business partners remain master-data entities.
    $permissionsByCode = @{}
    foreach ($permission in @(Get-Records 'security/permissions')) {
        $permissionsByCode[$permission.code] = $permission
    }
    $roleSpecs = @{
        WHADMIN = @{
            name = 'Warehouse Supervisor'
            description = 'Approves and controls SIS warehouse operations; does not perform billing.'
            permissions = @(
                'MASTER.READ', 'MASTER.WRITE', 'REPORTING.READ',
                'INBOUND.READ', 'INBOUND.APPROVE', 'INBOUND.ASSIGN', 'INBOUND.QUARANTINE_DISPOSE', 'INBOUND.CANCEL',
                'INVENTORY.READ', 'INVENTORY.IDENTITY', 'INVENTORY.MOVE', 'INVENTORY.STATUS_CHANGE', 'INVENTORY.ADJUST', 'INVENTORY.COUNT', 'INVENTORY.TRANSFER',
                'OUTBOUND.READ', 'OUTBOUND.TRANSPORT', 'OUTBOUND.SHIP', 'OUTBOUND.DELIVER', 'OUTBOUND.CANCEL', 'OUTBOUND.CONFIG'
            )
        }
        WAREHOUSE_PLANNER = @{
            name = 'Warehouse Planner'
            description = 'Creates inbound and outbound plans without approval, execution, or cancellation authority.'
            permissions = @('MASTER.READ', 'REPORTING.READ', 'INVENTORY.READ', 'INBOUND.READ', 'INBOUND.PLAN', 'OUTBOUND.READ', 'OUTBOUND.PLAN')
        }
        RECEIVER = @{
            name = 'Receiver Warehouse'
            description = 'Receives and puts away stock without approval, QC, disposal, or cancellation authority.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'INVENTORY.IDENTITY', 'INBOUND.READ', 'INBOUND.RECEIVE', 'INBOUND.PUTAWAY')
        }
        PICKER = @{
            name = 'Picker Warehouse'
            description = 'Picks and stages outbound stock.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'OUTBOUND.READ', 'OUTBOUND.PICK', 'OUTBOUND.STAGE')
        }
        IC = @{
            name = 'Inventory Control'
            description = 'Performs inventory identity, movement, status, count, adjustment, and transfer operations.'
            permissions = @('MASTER.READ', 'INBOUND.READ', 'OUTBOUND.READ', 'REPORTING.READ', 'INVENTORY.READ', 'INVENTORY.IDENTITY', 'INVENTORY.MOVE', 'INVENTORY.STATUS_CHANGE', 'INVENTORY.ADJUST', 'INVENTORY.COUNT', 'INVENTORY.TRANSFER')
        }
        QC = @{
            name = 'Quality Control'
            description = 'Records quality inspections without quarantine-disposition authority.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'INBOUND.READ', 'INBOUND.QC')
        }
        REWORK_OPERATOR = @{
            name = 'Rework Operator'
            description = 'Executes assigned quarantine rework without QC or disposition authority.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'INBOUND.READ', 'INBOUND.REWORK')
        }
        PACKER = @{
            name = 'Outbound Checker and Packer'
            description = 'Checks and packs picked stock without shipment or cancellation authority.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'OUTBOUND.READ', 'OUTBOUND.CHECK', 'OUTBOUND.PACK')
        }
        DISPATCHER = @{
            name = 'Warehouse Dispatcher'
            description = 'Dispatches shipments and records delivery execution.'
            permissions = @('MASTER.READ', 'INVENTORY.READ', 'OUTBOUND.READ', 'OUTBOUND.SHIP', 'OUTBOUND.DELIVER')
        }
        BILLING_OPERATOR = @{
            name = 'Billing Operator'
            description = 'Configures and prepares billing without approval, issue, or payment authority.'
            permissions = @('MASTER.READ', 'REPORTING.READ', 'BILLING.READ', 'BILLING.CONFIGURE', 'BILLING.PREPARE')
        }
        BILLING_APPROVER = @{
            name = 'Billing Approver'
            description = 'Approves billing and issues invoices or credit notes without payment-entry authority.'
            permissions = @('REPORTING.READ', 'BILLING.READ', 'BILLING.APPROVE', 'BILLING.ISSUE')
        }
        CASHIER = @{
            name = 'SIS Cashier'
            description = 'Records invoice payments without billing preparation or approval authority.'
            permissions = @('REPORTING.READ', 'BILLING.READ', 'BILLING.PAYMENT')
        }
    }
    $roles = @{}
    foreach ($code in $roleSpecs.Keys) {
        $spec = $roleSpecs[$code]
        $roles[$code] = Ensure-Role $code $spec.name $spec.description $spec.permissions $permissionsByCode
    }
    $accounts = @{}
    foreach ($spec in @(
        @('whm', 'WHADMIN'), @('receiver.warehouse', 'RECEIVER'),
        @('picker.warehouse', 'PICKER'), @('ic.warehouse', 'IC')
    )) {
        $matches = @(Get-Records 'security/accounts' | Where-Object username -EQ $spec[0])
        if ($matches.Count -ne 1) { throw "Expected existing account: $($spec[0])" }
        $accounts[$spec[0]] = $matches[0]
    }
    $activeStatus = Require-One 'security/account-statuses' 'ACTIVE'
    foreach ($spec in @(
        @('whm', 'WHADMIN'), @('receiver.warehouse', 'RECEIVER'),
        @('picker.warehouse', 'PICKER'), @('ic.warehouse', 'IC')
    )) {
        Ensure-AccountSetup $accounts[$spec[0]] $roles[$spec[1]] $owner $warehouse $activeStatus
    }
    $newAccountSpecs = @(
        @('planner.warehouse', 'Planner Warehouse', 'WAREHOUSE_PLANNER'),
        @('qc.warehouse', 'QC Warehouse', 'QC'),
        @('rework.warehouse', 'Rework Warehouse', 'REWORK_OPERATOR'),
        @('packer.warehouse', 'Packer Warehouse', 'PACKER'),
        @('dispatcher.warehouse', 'Dispatcher Warehouse', 'DISPATCHER'),
        @('billing.operator', 'SIS Billing Operator', 'BILLING_OPERATOR'),
        @('billing.approver', 'SIS Billing Approver', 'BILLING_APPROVER'),
        @('cashier.sis', 'SIS Cashier', 'CASHIER')
    )
    foreach ($spec in $newAccountSpecs) {
        Ensure-OperationalAccount $spec[0] $spec[1] $roles[$spec[2]] $owner $warehouse $activeStatus | Out-Null
    }

    # The manifest never includes the bearer token. Temporary passwords appear
    # only for accounts created during this run so an administrator can deliver
    # them once; reruns return an empty new_credentials list.
    [pscustomobject]@{
        created = $script:created
        updated = $script:updated
        reused = $script:reused
        owner = @{ id = $ownerId; code = $owner.code; name = $owner.name }
        warehouse = @{ id = $warehouseId; code = $warehouse.code; name = $warehouse.name }
        locations_total = @(Get-Records "master/warehouses/$warehouseId/locations").Count
        categories_total = @(Get-Records "master/item-categories?owner_id=$ownerId").Count
        items_total = @(Get-Records "master/items?owner_id=$ownerId").Count
        items_packaged = @($items | Where-Object { $itemPackByCode.ContainsKey($_.code) }).Count
        business_partners_typed = $partnerTypeMap.Count
        picking_strategies = @('SIS_BDG_FEFO', 'SIS_BDG_FIFO', 'SIS_BDG_LOCATION', 'SIS_BDG_LOT')
        putaway_strategy = 'SIS_BDG_STANDARD'
        activated_accounts = @(
            'whm', 'planner.warehouse', 'receiver.warehouse', 'picker.warehouse', 'ic.warehouse',
            'qc.warehouse', 'rework.warehouse', 'packer.warehouse', 'dispatcher.warehouse',
            'billing.operator', 'billing.approver', 'cashier.sis'
        )
        new_credentials = $script:newCredentials
    }
} finally {
    if ($ownedSession) {
        try { Invoke-Api POST 'auth/logout' | Out-Null }
        catch { Write-Warning 'Could not revoke the seed login session. Revoke it manually.' }
    }
    $Token = $null
    $headers.Clear()
}
