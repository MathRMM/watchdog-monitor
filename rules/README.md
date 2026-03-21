# .planning — Rotina de Planejamento de Código

Este diretório contém a rotina de desenvolvimento assistido por IA utilizada neste projeto. Todo desenvolvimento significativo passa obrigatoriamente por três etapas sequenciais antes de qualquer código ser escrito.

---

## Por que essa rotina existe

LLMs perdem contexto, inventam soluções e pulam etapas quando não há estrutura. Esta rotina resolve isso dividindo o trabalho em três responsabilidades bem separadas:

- **PRD** → o que precisa ser feito e por quê
- **SPEC** → como será feito e em que ordem
- **Implementação** → execução pura, sem decisão

Nenhuma etapa avança sem a anterior estar aprovada.

---

## Visão geral do fluxo

```
 Ideia / Solicitação
        │
        ▼
┌───────────────┐
│   1. PRD      │  Levantamento de requisitos
│               │  Saída: prd_{alteracao}.md
└───────┬───────┘
        │ aprovado pelo usuário
        ▼
┌───────────────┐
│   2. SPEC     │  Plano técnico em fases
│               │  Saída: spec_{alteracao}.md
└───────┬───────┘
        │ aprovado pelo usuário
        ▼
┌───────────────────────────────┐
│   3. IMPLEMENTAÇÃO            │
│                               │
│   Fase 1 → Fase 2 → Fase N   │  Uma fase por vez
│                               │  Saída: Makefile + commits
└───────────────────────────────┘
        │ bloqueio?
        ▼
┌───────────────┐
│   issue.md    │  Documenta o problema
│               │  Aguarda instrução
└───────────────┘
```

---

## Estrutura de arquivos

```
.planning/
├── README.md                          ← este arquivo
│
├── prd_{alteracao}.md                 ← gerado na etapa 1
├── spec_{alteracao}.md                ← gerado na etapa 2
│
└── issues/
    └── issue_{fase}_{descricao}.md   ← gerado apenas em caso de bloqueio
```

Fora do `.planning/`, na raiz do projeto:

```
Makefile                               ← criado na Fase 1, atualizado em toda fase
.env.example                           ← chaves sem valores, commitado
.env.local                             ← ambiente local, no .gitignore
.env.hml                               ← homologação, no .gitignore
.env.prod                              ← produção, no .gitignore
```

---

## As três etapas em detalhe

### Etapa 1 — PRD `prd_{alteracao}.md`

**Responsabilidade:** entender o problema antes de pensar em solução.

O que o PRD documenta:
- Contexto e problema a ser resolvido
- Requisitos funcionais (o que o sistema deve fazer)
- Requisitos não funcionais (performance, compatibilidade, disponibilidade)
- Regras de negócio (o que é permitido, proibido ou obrigatório)
- Requisitos de segurança
- Casos de uso, fluxos alternativos e casos de borda
- O que está **fora** do escopo
- Critérios de aceitação

**Regra:** nenhum código é escrito nesta etapa. Lacunas são marcadas como `[PENDENTE]` — nunca inventadas.

**Como ativar:** descreva a funcionalidade ou alteração para o Claude Code. Ele conduzirá as perguntas necessárias e gerará o arquivo.

---

### Etapa 2 — SPEC `spec_{alteracao}.md`

**Responsabilidade:** transformar requisitos em um plano executável.

O que o SPEC documenta:
- Tecnologias e versões envolvidas
- Mapa de arquivos: `[NEW]` criados, `[MOD]` modificados, `[DEL]` removidos, `[REF]` consultados
- Fases de implementação sugeridas pelo Claude com base no PRD
- Para cada fase: objetivo, arquivos, o que fazer, testes TDD e checklist de validação
- Contratos e interfaces compartilhados entre fases
- Rastreabilidade completa: cada RF, RN e SEC do PRD tem uma fase correspondente

**Regra:** o Claude sugere as fases e aguarda aprovação antes de detalhar o SPEC completo. Nenhuma fase avança sem o checklist da anterior estar 100% marcado.

**Como ativar:** com o PRD aprovado, diga ao Claude Code "avançar para o SPEC".

---

### Etapa 3 — Implementação

**Responsabilidade:** execução pura, sem decisão.

O que a etapa de implementação faz:
- Executa **uma fase por vez** conforme o SPEC
- Segue a ordem TDD: escreve testes → confirma falha → implementa → confirma verde → refatora
- Atualiza o Makefile em toda fase
- Realiza commit atômico ao final de cada fase com mensagem padronizada
- Para completamente e gera um `issue.md` se qualquer bloqueio for encontrado

**Regra:** o Claude Code é agnóstico de decisão nesta etapa. Ele não sugere arquiteturas alternativas nem adiciona funcionalidades fora do SPEC. Se algo está errado no plano, o caminho é o issue — nunca uma gambiarra.

**Como ativar:** com o SPEC aprovado, diga "execute a Fase 1", "Fase 2", etc.

---

## Convenções de nomenclatura

| Arquivo | Padrão | Exemplo |
|---|---|---|
| PRD | `prd_{alteracao}.md` | `prd_autenticacao_oauth.md` |
| SPEC | `spec_{alteracao}.md` | `spec_autenticacao_oauth.md` |
| Issue | `issue_{fase}_{descricao}.md` | `issue_2_migration_falhou.md` |
| Branch | `feature/{alteracao}` | `feature/autenticacao-oauth` |

O campo `{alteracao}` deve ser em `snake_case`, descritivo e curto.

---

## Padrão de commits

```
{tipo}({escopo}): {descrição curta}

Fase {N} — {nome da fase}
Refs: RF{nn}, RN{nn}
```

Tipos:

| Tipo | Uso |
|---|---|
| `feat` | Nova funcionalidade |
| `fix` | Correção de bug |
| `test` | Adição ou correção de testes |
| `chore` | Configuração, build, Makefile, docker |
| `refactor` | Refatoração sem mudança de comportamento |
| `docs` | Documentação |

---

## Branches

```
main          → produção, nunca commitar direto
develop       → integração
feature/{x}   → branch da alteração, criada a partir de develop
```

Ao concluir todas as fases, abrir Pull Request de `feature/{alteracao}` → `develop` e revisar os critérios de aceitação do PRD antes de aprovar.

---

## Makefile — comandos disponíveis

O Makefile é criado na Fase 1 e incrementado a cada fase. Para ver todos os comandos disponíveis no estado atual do projeto:

```bash
make help
```

Grupos de comandos cobertos:

| Grupo | Comandos |
|---|---|
| Setup | `make setup` |
| Desenvolvimento | `make dev`, `make start`, `make stop` |
| Testes | `make test`, `make test-watch` |
| Banco de dados | `make migrate`, `make seed` |
| Build e deploy | `make build`, `make deploy` |
| Docker | `make docker-up`, `make docker-down`, `make docker-logs` |
| Docker por ambiente | `make docker-up-hml`, `make docker-up-prod` |

Variáveis de ambiente nunca são hardcoded no Makefile. Cada ambiente lê seu próprio `.env.*`.

---

## Quando um bloqueio ocorre

Se durante a implementação o Claude Code encontrar um problema que não consegue resolver dentro do plano, ele irá:

1. Parar toda implementação imediatamente
2. Gerar `.planning/issues/issue_{fase}_{descricao}.md` com o problema documentado
3. Apresentar hipóteses e opções de resolução — sem escolher
4. Aguardar instrução

O arquivo de issue documenta: o problema, o contexto, a evidência (erro/log), o impacto, hipóteses de causa e opções de resolução com trade-offs.

**Nunca** retome a implementação sem revisar o issue e dar uma instrução explícita.

---

## Checklist rápido antes de começar

- [ ] Descrevi claramente o que precisa ser feito?
- [ ] O PRD foi gerado e aprovado?
- [ ] O SPEC foi gerado, fases revisadas e aprovado?
- [ ] Estou executando **uma fase por vez**?
- [ ] O checklist da fase anterior está 100% marcado?
- [ ] Os arquivos `.env.*` estão no `.gitignore`?
