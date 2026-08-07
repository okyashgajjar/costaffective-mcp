$ProgressPreference = 'SilentlyContinue'
if (-not (Test-Path "w64devkit")) {
    Write-Host "Downloading portable 64-bit GCC (w64devkit)..."
    Invoke-WebRequest -Uri "https://github.com/skeeto/w64devkit/releases/download/v1.23.0/w64devkit-1.23.0.zip" -OutFile "w64devkit.zip"
    Write-Host "Extracting w64devkit (this might take a minute)..."
    Expand-Archive -Path "w64devkit.zip" -DestinationPath "." -Force
}

# Add GCC to PATH for this script session
$env:PATH = "$PWD\w64devkit\bin;" + $env:PATH

Write-Host "Building CostWise with CGO_ENABLED=1..."
$env:CGO_ENABLED = "1"
go build -o costwise.exe ./cmd/costwise/

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful! Running benchmark..."
    python tests/benchmark_token_efficiency.py
} else {
    Write-Host "Build failed."
    exit 1
}
