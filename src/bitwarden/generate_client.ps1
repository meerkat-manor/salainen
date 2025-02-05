

$openApiFile = "./vault-management-api.json"

$codeGen = $env:USERPROFILE + "/go/pkg/mod/github.com/oapi-codegen/oapi-codegen@v2.3.0/cmd/oapi-codegen/oapi-codegen.go"

#go get github.com/oapi-codegen/oapi-codegen/pkg/codegen@v2.3.0
# for the binary install
#go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

Write-Host "Generating code from $openApiFile"

#go run $codeGen  --config ../build/types.cfg.yaml $openApiFile 
#go run $codeGen  --config ../build/server.cfg.yaml $openApiFile

oapi-codegen --config ./types.cfg.yaml $openApiFile
oapi-codegen --config ./client.cfg.yaml $openApiFile


