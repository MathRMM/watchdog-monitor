# Rule: SPEC — Especificação Técnica e Plano de Implementação

## Quando esta rule é ativada

Ative esta rule **somente após o PRD estar aprovado**. O gatilho é o usuário dizer algo como "avançar para o SPEC", "gerar o SPEC" ou "próxima etapa".

O arquivo `.planning/prd_{alteracao}.md` **deve existir** antes de começar. Se não existir, interrompa e oriente o usuário a executar a etapa de PRD primeiro.

---

## Objetivo

Transformar os requisitos do PRD em um plano técnico executável, dividido em **fases sequenciais**. Cada fase é pequena o suficiente para ser implementada em um único contexto de IA, com critérios claros de validação e testes TDD definidos antes da implementação.

Você **não escreve código de produção** nesta etapa. Você mapeia, planeja e estrutura.

---

## Comportamento esperado

### 1. Leitura do PRD

Leia o arquivo `.planning/prd_{alteracao}.md` na íntegra antes de qualquer coisa. Extraia:

- Os requisitos funcionais (RF) e não funcionais (RNF)
- As regras de negócio (RN)
- Os critérios de aceitação
- Os casos de borda documentados

### 2. Mapeamento de arquivos

Antes de definir as fases, faça o levantamento completo dos arquivos envolvidos:

- **Arquivos existentes** que serão modificados → marcar com `[MOD]`
- **Arquivos novos** que serão criados → marcar com `[NEW]`
- **Arquivos removidos** → marcar com `[DEL]`
- **Arquivos lidos/consultados mas não alterados** → marcar com `[REF]`

### 3. Sugestão de fases

Com base no PRD, sugira as fases de implementação e **aguarde aprovação do usuário** antes de detalhar cada uma. Apresente as fases em formato de lista com uma linha descrevendo o objetivo de cada uma.

Exemplo de apresentação:

> "Sugiro dividir a implementação em 4 fases. Confirme para eu detalhar o SPEC completo:
> - **Fase 1 — Infraestrutura:** configuração de ambiente, containers e variáveis
> - **Fase 2 — Modelo de dados:** criação das entidades e migrations
> - **Fase 3 — Lógica de negócio:** serviços, validações e regras
> - **Fase 4 — Exposição:** controllers, rotas e contratos de API"

### 4. Geração do SPEC

Após aprovação das fases, gere o arquivo `.planning/spec_{alteracao}.md` seguindo a estrutura abaixo.

---

## Estrutura do arquivo `spec_{alteracao}.md`

```markdown
# SPEC — {Nome da Alteração}

**PRD de referência:** `.planning/prd_{alteracao}.md`
**Data:** {data de criação}
**Status:** draft | revisão | aprovado
**Autor:** Claude Code

---

## 1. Visão Geral Técnica

> Resumo da abordagem técnica escolhida para atender os requisitos do PRD.
> Inclua: padrões arquiteturais adotados, justificativa das escolhas, e qualquer
> decisão técnica relevante que impacte todas as fases.

## 2. Tecnologias e Versões

| Tecnologia     | Versão   | Uso na alteração                        |
|----------------|----------|-----------------------------------------|
| {ex: Node.js}  | {18.x}   | {Runtime principal}                     |
| {ex: Postgres} | {15}     | {Persistência dos dados de X}           |
| {ex: Redis}    | {7}      | {Cache de sessão}                       |

> Liste apenas as tecnologias **diretamente envolvidas** nesta alteração.

## 3. Mapa de Arquivos

{raiz do projeto}
├── [REF] arquivo-existente-consultado.ts
├── [MOD] arquivo-existente-modificado.ts
├── [NEW] novo-arquivo-criado.ts
└── [DEL] arquivo-a-ser-removido.ts


Legenda:
- `[NEW]` — Arquivo criado nesta alteração
- `[MOD]` — Arquivo existente que será modificado
- `[DEL]` — Arquivo que será removido
- `[REF]` — Arquivo consultado, mas não alterado

> Para cada arquivo `[NEW]` e `[MOD]`, descreva em uma linha sua responsabilidade:
> - `[NEW] src/services/auth.service.ts` — Lógica de autenticação OAuth, expõe `login()` e `callback()`
> - `[MOD] src/app.module.ts` — Registrar AuthModule no módulo raiz

---

## 4. Fases de Implementação

> Cada fase deve ser pequena o suficiente para ser executada em um único contexto.
> A fase só avança quando todos os itens do checklist estiverem marcados.

---

### Fase {N} — {Nome da Fase}

**Objetivo:** {Uma frase descrevendo o que esta fase entrega ao final.}

**Arquivos desta fase:**
- `[NEW/MOD]` `caminho/do/arquivo.ts` — o que ele faz nesta fase

**O que fazer:**
> Descrição em linguagem natural do que deve ser implementado nesta fase.
> Sem pseudocódigo, sem implementação — apenas o que e por quê.
> Referencie os RFs, RNs e RNFs do PRD que esta fase atende.
> Ex: "Criar o serviço de autenticação que atende RF03 e aplica a regra RN02."

**Testes (TDD):**

> Descreva os testes que devem ser escritos **antes** da implementação desta fase.
> Use linguagem de alto nível — o que deve ser validado, não como implementar.

- [ ] `{contexto}` → dado {condição}, quando {ação}, então {resultado esperado}
- [ ] `{contexto}` → dado {condição}, quando {ação}, então {resultado esperado}
- [ ] Caso de borda: {descrição do cenário limite a ser coberto}

**Checklist de validação:**

- [ ] {O comportamento X funciona conforme RF0N}
- [ ] {A regra de negócio RN0N é respeitada}
- [ ] {O requisito não funcional RNF0N é atendido}
- [ ] {Todos os testes desta fase estão passando}
- [ ] {Nenhum teste de fases anteriores foi quebrado}

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. Somente então iniciar a Fase {N+1}.
```
---

## 5. Ordem de Execução e Dependências entre Fases

```
Fase 1 → Fase 2 → Fase 3 → Fase 4
           ↑
     depende de Fase 1
```

> Descreva dependências não óbvias entre fases. Se uma fase pode ser executada
> em paralelo com outra, sinalize aqui.

## 6. Contratos e Interfaces

> Defina os contratos que as fases irão implementar: tipos, interfaces, schemas,
> contratos de API (request/response). Isso serve como referência compartilhada
> entre fases para evitar retrabalho.

```typescript
// Exemplo: contrato definido aqui, implementado nas fases
interface ExemploContrato {
  campo: tipo;
}
```

> Se não houver contratos a definir, documente: "Sem contratos compartilhados entre fases."

## 7. Riscos Técnicos

| Risco                                      | Fase afetada | Mitigação                              |
|--------------------------------------------|--------------|----------------------------------------|
| {ex: migration pode quebrar dados legados} | Fase 2       | {Criar backup antes de executar}       |
| {ex: dependência externa pode estar fora}  | Fase 1       | {Validar healthcheck antes de avançar} |

## 8. Rastreabilidade PRD → SPEC

| Requisito PRD | Atendido na fase | Observação                        |
|---------------|------------------|-----------------------------------|
| RF01          | Fase 2           | —                                 |
| RF02          | Fase 3           | Depende de RF01 estar implementado|
| RN01          | Fase 3           | Validado no checklist da fase     |
| SEC01         | Fase 1           | —                                 |

> Todos os RFs, RNs e SECs do PRD devem aparecer nesta tabela.
> Se algum requisito não tiver fase correspondente, sinalize como `[PENDENTE]`.

---

## Histórico de Revisões

| Versão | Data   | Alteração        |
|--------|--------|------------------|
| 1.0    | {data} | Criação inicial  |
```

---

## Regras de conduta desta etapa

1. **Nunca escreva código de produção.** Contratos e interfaces são exceção — podem ser esboçados na seção 6 como referência entre fases, mas não implementados.

2. **Toda fase deve referenciar o PRD.** Cada "O que fazer" deve citar ao menos um RF, RN ou RNF. Se uma fase não consegue referenciar nada do PRD, questione se ela é necessária.

3. **Fases devem ser atômicas.** Uma fase não pode depender de outra fase que ainda não foi concluída dentro do mesmo SPEC. A ordem de execução deve ser linear e explícita.

4. **Testes antes da implementação.** Os testes descritos em cada fase são escritos **antes** do código. A fase de implementação usará este SPEC como referência para saber o que testar primeiro.

5. **Checklist é o critério de avanço.** A etapa de implementação não avança para a próxima fase enquanto o checklist da fase atual não estiver 100% marcado.

6. **Rastreabilidade total.** Nenhum requisito do PRD pode ficar sem fase correspondente. A tabela da seção 8 deve cobrir todos os RFs, RNs e SECs.

7. **Sinalizar lacunas com `[PENDENTE]`.** Se uma decisão técnica não puder ser tomada agora, documente como `[PENDENTE: motivo]` e não invente uma solução.

---
```
## Output esperado

```
.planning/
├── prd_{alteracao}.md    ← gerado na etapa anterior (não modificar)
└── spec_{alteracao}.md   ← gerado nesta etapa
```

---

## Transição para próxima etapa

Após gerar o SPEC e o usuário aprovar, informe:

> "SPEC gerado em `.planning/spec_{alteracao}.md`. Quando quiser iniciar, me diga **qual fase executar** — trabalharemos uma fase por vez para manter o contexto controlado."
