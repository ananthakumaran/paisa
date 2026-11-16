[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [ValidateNotNullOrEmpty()]
    [string[]]$Path,

    [Parameter(Mandatory)]
    [ValidateSet('test-signing', 'release-signing')]
    [string]$SigningPolicy,

    [string]$ExpectedThumbprint
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$allowedStatus = @('Valid')
if ($SigningPolicy -eq 'test-signing') {
    # The Foundation's self-signed test certificate is not trusted by Windows.
    # Do not install it as a trusted root on the runner or on users' machines.
    $allowedStatus += 'NotTrusted', 'UnknownError'
}

foreach ($file in $Path) {
    $signature = Get-AuthenticodeSignature -LiteralPath $file
    if (-not $signature.SignerCertificate -or $signature.Status.ToString() -notin $allowedStatus) {
        throw "$file signature is $($signature.Status): $($signature.StatusMessage)"
    }

    $thumbprint = $signature.SignerCertificate.Thumbprint
    if ([string]::IsNullOrWhiteSpace($ExpectedThumbprint)) {
        $ExpectedThumbprint = $thumbprint
    }
    if ($thumbprint -ne $ExpectedThumbprint) {
        throw "$file was signed with a different certificate: $thumbprint"
    }
    if (-not $signature.TimeStamperCertificate) {
        if ($SigningPolicy -eq 'release-signing') {
            throw "$file has no Authenticode timestamp"
        }
        Write-Warning "$file test signature has no Authenticode timestamp"
    }

    Write-Host "$file : $($signature.Status), signer=$($signature.SignerCertificate.Subject), thumbprint=$thumbprint"
}

if ($env:GITHUB_OUTPUT) {
    "certificate-thumbprint=$ExpectedThumbprint" >> $env:GITHUB_OUTPUT
}
