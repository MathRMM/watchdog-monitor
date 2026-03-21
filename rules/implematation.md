# Rule: IMPLEMENTAÇÃO — Execução por Fase

## Quando esta rule é ativada

Ative esta rule quando o usuário disser qual fase executar, ex: "execute a Fase 1", "vamos implementar a Fase 2", "próxima fase".

O arquivo `.planning/spec_{alteracao}.md` **deve existir e estar aprovado** antes de começar. Se não existir, interrompa e oriente o usuário a executar a etapa de SPEC primeiro.

---

## Objetivo

Executar **uma fase por vez** do SPEC de forma disciplinada: testes primeiro, código depois, Makefile sempre atualizado, e git em cada entrega atômica. Em caso de bloqueio, parar completamente e documentar o problema — nunca improvisar uma solução não planejada.

---

## Comportamento esperado

### 1. Antes de iniciar qualquer fase

Leia **obrigatoriamente**:
- `.planning/spec_{alteracao}.md` — para seguir o plano da fase atual
- O checklist da fase anterior (se houver) — confirmar que está 100% marcado antes de avançar

Se o checklist da fase anterior tiver itens não marcados, **interrompa** e informe o usuário. Não inicie a próxima fase com pendências abertas.

### 2. Ordem de execução dentro de cada fase

Siga esta ordem estritamente. Não pule etapas.

```
1. Escrever os testes (TDD)
2. Confirmar que os testes falham (red)
3. Implementar o código
4. Confirmar que os testes passam (green)
5. Refatorar se necessário, mantendo testes verdes
6. Atualizar o Makefile
7. Commit git
8. Marcar checklist da fase
```

### 3. Gestão do contexto

Esta rule é **agnóstica de decisão**. Isso significa:

- Não sugira arquiteturas alternativas
- Não questione escolhas do SPEC
- Não adicione funcionalidades não planejadas
- Não refatore código fora do escopo da fase atual
- Se identificar algo problemático no SPEC durante a implementação → **bloqueie e documente** em vez de improvisar

---

## Regras de Git

### Branches

```
main / master          — produção, nunca commitar direto
develop                — branch de integração
feature/{alteracao}    — branch desta alteração, criada a partir de develop
```

Ao iniciar a implementação de uma alteração, verifique se a branch `feature/{alteracao}` existe. Se não existir, crie:

```bash
git checkout develop
git pull origin develop
git checkout -b feature/{alteracao}
```

### Commits por fase

Cada fase gera **ao menos um commit atômico** ao final. O padrão de mensagem é:

```
{tipo}({escopo}): {descrição curta}

Fase {N} — {nome da fase}
Refs: RF{nn}, RN{nn}  ← referenciar os itens do PRD atendidos
```

Tipos permitidos:
- `feat` — nova funcionalidade
- `fix` — correção de bug
- `test` — adição ou correção de testes
- `chore` — configuração, build, Makefile, docker
- `refactor` — refatoração sem mudança de comportamento
- `docs` — documentação

Exemplos:
```
chore(infra): adicionar docker-compose para postgres e redis

Fase 1 — Ambiente de container
Refs: RNF01, RNF02

feat(auth): implementar serviço de autenticação OAuth

Fase 3 — Lógica de negócio
Refs: RF03, RN02
```

### Regras adicionais de git

- **Nunca force push** em `develop` ou `main`
- **Nunca commitar** arquivos `.env`, segredos ou credenciais
- **Sempre** incluir o `Makefile` atualizado no commit da fase
- Se a fase tiver subtarefas grandes, pode-se fazer commits intermediários com o sufixo `[wip]`, mas o commit final da fase não deve ter `[wip]`

---

## Makefile

O Makefile deve ser criado na Fase 1 e **atualizado em toda fase subsequente**. Ele é incluído no commit de cada fase.

### Estrutura padrão

```makefile
# =============================================================================
# {NOME DO PROJETO}
# =============================================================================

.PHONY: help setup dev start stop test test-watch \
        migrate seed build deploy \
        docker-up docker-down docker-logs \
        docker-up-hml docker-up-prod

# Exibe os comandos disponíveis
help:
	@echo ""
	@echo "  Comandos disponíveis:"
	@echo ""
	@echo "  Setup"
	@echo "    make setup          Instala dependências"
	@echo ""
	@echo "  Desenvolvimento"
	@echo "    make dev            Inicia em modo desenvolvimento"
	@echo "    make start          Inicia em modo produção"
	@echo "    make stop           Para a aplicação"
	@echo ""
	@echo "  Testes"
	@echo "    make test           Executa todos os testes"
	@echo "    make test-watch     Executa testes em modo watch"
	@echo ""
	@echo "  Banco de dados"
	@echo "    make migrate        Executa migrations pendentes"
	@echo "    make seed           Popula banco com dados iniciais"
	@echo ""
	@echo "  Build e Deploy"
	@echo "    make build          Gera build de produção"
	@echo "    make deploy         Realiza deploy"
	@echo ""
	@echo "  Docker"
	@echo "    make docker-up      Sobe containers (ambiente local)"
	@echo "    make docker-down    Derruba containers"
	@echo "    make docker-logs    Exibe logs dos containers"
	@echo "    make docker-up-hml  Sobe containers (homologação)"
	@echo "    make docker-up-prod Sobe containers (produção)"
	@echo ""

# -----------------------------------------------------------------------------
# Setup
# -----------------------------------------------------------------------------
setup:
	{comando de instalação de dependências}

# -----------------------------------------------------------------------------
# Desenvolvimento
# -----------------------------------------------------------------------------
dev:
	{comando para iniciar em modo dev}

start:
	{comando para iniciar em modo produção}

stop:
	{comando para parar}

# -----------------------------------------------------------------------------
# Testes
# -----------------------------------------------------------------------------
test:
	{comando para rodar todos os testes}

test-watch:
	{comando para rodar testes em modo watch}

# -----------------------------------------------------------------------------
# Banco de dados
# -----------------------------------------------------------------------------
migrate:
	{comando para rodar migrations}

seed:
	{comando para rodar seeds}

# -----------------------------------------------------------------------------
# Build e Deploy
# -----------------------------------------------------------------------------
build:
	{comando de build}

deploy:
	{comando de deploy}

# -----------------------------------------------------------------------------
# Docker — ambientes
# -----------------------------------------------------------------------------
docker-up:
	docker compose --env-file .env.local up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-up-hml:
	docker compose --env-file .env.hml up -d

docker-up-prod:
	docker compose --env-file .env.prod up -d
```

### Regras do Makefile

- **Todo comando novo criado em uma fase** deve ter um alvo correspondente no Makefile
- Alvos que ainda não têm implementação recebem um placeholder com aviso:
  ```makefile
  deploy:
  	@echo "[TODO] Comando de deploy não configurado ainda"
  ```
- O alvo `help` deve sempre estar atualizado com todos os comandos disponíveis
- Variáveis de ambiente **nunca** são hardcoded no Makefile — sempre lidas de `.env.*`
- Arquivos `.env.local`, `.env.hml`, `.env.prod` devem estar no `.gitignore`; commitar apenas `.env.example` com as chaves sem valores

---

## Fluxo de Fallback — Bloqueio por Falha

Quando qualquer um dos cenários abaixo ocorrer, **pare imediatamente** a implementação:

### Gatilhos de bloqueio

- Teste que deveria passar continua falhando após a implementação
- Erro não previsto no SPEC que impede a continuação da fase
- Dependência externa indisponível ou incompatível
- Conflito entre o que o SPEC pede e o estado real do código
- Ambiguidade que exige uma decisão de arquitetura não planejada
- Quebra de testes de fases anteriores causada pela fase atual

### Ação ao bloquear

1. **Pare toda implementação** — não tente contornar o problema
2. Reverta alterações não commitadas se o estado atual estiver inconsistente
3. Gere o arquivo `.planning/issues/issue_{fase}_{descricao_curta}.md`
4. Informe o usuário com a mensagem padrão abaixo

**Mensagem de bloqueio:**
> "⚠️ Implementação bloqueada na **Fase {N} — {nome}**.
> O problema foi documentado em `.planning/issues/issue_{fase}_{descricao}.md`.
> Revise o documento e me instrua sobre como proceder."

---

## Estrutura do arquivo `issue_{fase}_{descricao}.md`

```markdown
# Issue — Fase {N}: {Descrição Curta}

**Data:** {data}
**Fase afetada:** Fase {N} — {nome da fase}
**Status:** bloqueado

---

## Problema identificado

> Descreva objetivamente o que aconteceu. Sem especulação — apenas fatos observáveis.

## Contexto

> O que estava sendo implementado quando o problema foi encontrado?
> Qual era o estado do sistema antes do bloqueio?

## Evidência

> Cole aqui o erro, log, output de teste ou comportamento inesperado que causou o bloqueio.

```
{stack trace, output de teste, mensagem de erro}
```

## Impacto

> O que não pode ser continuado por causa deste problema?
> Quais fases subsequentes estão bloqueadas?

## Hipóteses

> Liste possíveis causas, sem afirmar qual é a correta.
> - Hipótese A: ...
> - Hipótese B: ...

## Opções de resolução

> Liste caminhos possíveis para desbloquear, com trade-offs de cada um.
> Não escolha — apresente as opções para decisão do usuário.

| Opção | Descrição                         | Trade-off                        |
|-------|-----------------------------------|----------------------------------|
| A     | {descrição}                       | {vantagem / desvantagem}         |
| B     | {descrição}                       | {vantagem / desvantagem}         |

## Arquivos envolvidos

- `caminho/do/arquivo.ts` — {relação com o problema}

---

**Aguardando instrução para retomar a implementação.**
```

---

## Checklist de encerramento de fase

Ao final de cada fase, antes de informar o usuário que a fase está completa, verifique:

- [ ] Todos os testes da fase estão passando
- [ ] Nenhum teste de fases anteriores foi quebrado
- [ ] O Makefile foi atualizado e incluído no commit
- [ ] O commit da fase foi realizado com a mensagem no padrão definido
- [ ] Todos os itens do checklist do SPEC para esta fase estão marcados
- [ ] Nenhum arquivo `.env` com credenciais reais foi commitado

Somente após todos os itens acima, informe:

> "✅ Fase {N} — {nome} concluída. Todos os testes passando, Makefile atualizado e commit realizado.
> Pronto para iniciar a **Fase {N+1} — {nome}** quando quiser."

Se for a última fase:

> "✅ Fase {N} — {nome} concluída. Todas as fases do SPEC foram implementadas.
> Recomendo revisar os critérios de aceitação do PRD antes de abrir o Pull Request de `feature/{alteracao}` → `develop`."

---

## Output esperado ao longo das fases

```
.planning/
├── prd_{alteracao}.md
├── spec_{alteracao}.md
└── issues/
    └── issue_{fase}_{descricao}.md   ← gerado apenas em caso de bloqueio

{raiz do projeto}/
└── Makefile                          ← criado na Fase 1, atualizado em toda fase
```
