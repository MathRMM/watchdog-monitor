# Issue — Fase 1: Go não instalado no ambiente

**Data:** 2026-03-21
**Fase afetada:** Fase 1 — Scaffold, Configuração e Windows Service
**Status:** bloqueado

---

## Problema identificado

Go não está instalado em nenhum ambiente acessível ao Claude Code (WSL nem Windows via interop). Sem Go, é impossível executar o ciclo TDD (red → green), inicializar o módulo com `go mod init`, baixar dependências com `go get`, e validar a compilação com `go build ./...`.

## Contexto

Ao iniciar a Fase 1, verificou-se que:
- `go version` falha com "command not found" no WSL
- `cmd.exe /c "where go"` retorna "não foi possível localizar"
- `powershell.exe /c "where go"` retorna "não foi possível localizar"
- Go não está em `/mnt/c/Program Files/Go/`, nem em `~/go/bin/`, nem em AppData

O projeto existe em `C:\Users\mathe\OneDrive\Documentos\Codigos\Watchdog Monitor\` mas nenhum toolchain Go está disponível.

## Evidência

```
$ go version
bash: go: command not found

$ cmd.exe /c "where go"
INFORMAÇÕES: não foi possível localizar arquivos para o(s) padrão(ões) especificado(s).

$ powershell.exe /c "where go"
Exit code 1 — comando não encontrado
```

## Impacto

- **Fase 1 completamente bloqueada**: impossível executar `go mod init`, `go get`, `go test`, `go build`
- **Fases 2–5 também bloqueadas**: dependem do toolchain Go
- Código-fonte pode ser escrito, mas ciclo TDD e checklist de validação não podem ser confirmados

## Hipóteses

- **Hipótese A**: Go nunca foi instalado nesta máquina
- **Hipótese B**: Go está instalado mas fora do PATH do Windows (ex: instalação manual sem adicionar ao PATH)
- **Hipótese C**: Go está acessível via algum gerenciador de versão (asdf, scoop, winget) mas não está no PATH padrão

## Opções de resolução

| Opção | Descrição | Trade-off |
|-------|-----------|-----------|
| A | Instalar Go no Windows via instalador oficial (go.dev/dl) — versão 1.22+ | Permanente, recomendado. Requer reiniciar o WSL após a instalação para que o PATH seja atualizado. |
| B | Instalar Go no WSL diretamente (`sudo apt install golang-go` ou via tarball oficial) | Permite rodar comandos Go direto no WSL onde Claude Code opera. Versão do apt pode ser antiga — preferir tarball 1.22+. |
| C | Escrever todos os arquivos de código agora e ter o usuário rodar `go mod tidy && go test ./internal/config/... && GOOS=windows GOARCH=amd64 go build ./...` manualmente no Windows após instalar Go | Avança a escrita do código, mas a validação TDD fica com o usuário. Aceitável apenas se o usuário preferir. |

## Arquivos envolvidos

- `go.mod` — precisa de `go mod init` para ser gerado corretamente
- `go.sum` — gerado por `go mod tidy`

---

**Aguardando instrução para retomar a implementação.**
