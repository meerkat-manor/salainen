
$openApiFile = "./bitwarden-management-api.json"

Write-Host "Generating code from $openApiFile"

$codeGen = $env:USERPROFILE + "/go/pkg/mod/github.com/oapi-codegen/oapi-codegen@v2.3.0/cmd/oapi-codegen/oapi-codegen.go"

#go get github.com/oapi-codegen/oapi-codegen/pkg/codegen@v2.3.0

#go run $codeGen  --config ../build/types.cfg.yaml $openApiFile 
#go run $codeGen  --config ../build/server.cfg.yaml $openApiFile

# for the binary install
# go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen --config ./types.cfg.yaml $openApiFile
oapi-codegen --config ./client.cfg.yaml $openApiFile


