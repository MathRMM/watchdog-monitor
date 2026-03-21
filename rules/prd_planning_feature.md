# Rule: PRD — Levantamento de Requisitos

## Quando esta rule é ativada

Ative esta rule quando o usuário descrever uma nova funcionalidade, alteração, correção significativa ou qualquer mudança que impacte comportamento do sistema. O gatilho pode ser uma descrição em linguagem natural, um ticket, ou uma instrução direta como "quero implementar X".

---

## Objetivo

Conduzir um levantamento profundo **antes de qualquer código ser escrito**. O resultado é um arquivo `prd_{alteracao}.md` salvo em `.planning/` que servirá como fonte de verdade para as etapas seguintes (SPEC e Implementação).

Você **não escreve código** nesta etapa. Você pergunta, analisa e documenta.

---

## Comportamento esperado

### 1. Entendimento inicial

Ao receber uma solicitação, faça perguntas claras para eliminar ambiguidades antes de começar o documento. Exemplos:

- "Qual o problema real que isso resolve?"
- "Existe algum comportamento atual que deve ser mantido?"
- "Há integrações externas envolvidas?"
- "Quem são os usuários afetados por essa mudança?"

Não avance para a escrita do PRD sem ter respostas suficientes. Se a solicitação for vaga, sinalize explicitamente o que está indefinido.

### 2. Geração do PRD

Após entender a solicitação, gere o arquivo `.planning/prd_{alteracao}.md` seguindo **exatamente** a estrutura abaixo.

---

## Estrutura do arquivo `prd_{alteracao}.md`

```markdown
# PRD — {Nome da Alteração}

**Data:** {data de criação}
**Status:** draft | revisão | aprovado

---

## 1. Contexto e Problema

> Descreva o cenário atual e o problema que esta alteração resolve.
> Seja específico: o que falha hoje? O que está ausente? Qual dor do usuário ou do sistema?

## 2. Objetivo

> O que esta alteração deve alcançar ao ser concluída?
> Use linguagem mensurável quando possível: "permitir que o usuário faça X sem precisar de Y".

## 3. Requisitos Funcionais

Liste cada comportamento esperado do sistema após a implementação.

| ID   | Requisito                                      | Prioridade     |
|------|------------------------------------------------|----------------|
| RF01 | Descrição clara do comportamento esperado      | Alta           |
| RF02 | ...                                            | Média          |

> Prioridades: Alta / Média / Baixa

## 4. Requisitos Não Funcionais

Liste restrições de qualidade que a solução deve respeitar.

| ID    | Requisito                                         | Categoria       |
|-------|---------------------------------------------------|-----------------|
| RNF01 | Tempo de resposta < 200ms para operações X        | Performance     |
| RNF02 | Compatível com Node.js 18+                        | Compatibilidade |
| RNF03 | ...                                               | ...             |

> Categorias sugeridas: Performance, Escalabilidade, Compatibilidade, Manutenibilidade, Disponibilidade

## 5. Regras de Negócio

Descreva as regras que governam o comportamento — não como o sistema funciona tecnicamente, mas o que é permitido, proibido ou obrigatório do ponto de vista do domínio.

- **RN01:** {descrição da regra}
- **RN02:** {descrição da regra}
- **RN03:** {exceções e casos especiais}

## 6. Requisitos de Segurança

Liste explicitamente os requisitos de segurança aplicáveis a esta alteração.

- **SEC01:** {ex: dados sensíveis não devem ser logados}
- **SEC02:** {ex: endpoint deve exigir autenticação JWT}
- **SEC03:** {ex: inputs devem ser sanitizados antes de persistência}

> Se não houver requisitos de segurança específicos, documente: "Nenhum requisito de segurança adicional identificado para esta alteração."

## 7. Casos de Uso

Descreva os fluxos principais e alternativos.

### Fluxo Principal
1. Usuário/sistema faz X
2. Sistema responde com Y
3. Estado final: Z

### Fluxos Alternativos
- **Alternativo A:** Se X não estiver disponível, então...
- **Alternativo B:** Em caso de falha em Y, o sistema deve...

### Casos de Borda
- {O que acontece com entradas vazias, nulas, ou inválidas?}
- {O que acontece em condições de concorrência?}
- {O que acontece se dependências externas falharem?}

## 8. Fora do Escopo

Liste explicitamente o que **não** será feito nesta alteração para evitar escopo deslizante.

- Não inclui migração de dados legados
- Não cobre o módulo X
- ...

## 9. Dependências e Riscos

| Tipo         | Descrição                                              | Impacto  |
|--------------|--------------------------------------------------------|----------|
| Dependência  | Serviço externo Y precisa estar disponível             | Alto     |
| Risco        | Mudança pode afetar comportamento do módulo Z          | Médio    |
| Risco        | ...                                                    | ...      |

## 10. Critérios de Aceitação

Liste as condições que precisam ser verdadeiras para esta alteração ser considerada completa e correta.

- [ ] {Comportamento A funciona conforme RF01}
- [ ] {Regra de negócio RN02 é respeitada em todos os cenários}
- [ ] {Requisito de segurança SEC01 foi implementado e validado}
- [ ] {Todos os casos de borda do item 7 têm tratamento definido}

---

## Histórico de Revisões

| Versão | Data       | Alteração                  |
|--------|------------|----------------------------|
| 1.0    | {data}     | Criação inicial            |
```

---

## Regras de conduta desta etapa

1. **Nunca escreva código.** Se sentir vontade de sugerir uma implementação, anote como observação no campo de dependências ou riscos, mas não implemente.

2. **Sinalize lacunas explicitamente.** Se uma seção não puder ser preenchida por falta de informação, escreva `[PENDENTE: descrição do que falta]` no campo correspondente. Não invente requisitos.

3. **Seja específico, não genérico.** Evite frases como "o sistema deve ser rápido". Prefira "o sistema deve responder em menos de 300ms para 95% das requisições sob carga de 100 usuários simultâneos".

4. **Preserve o que já existe.** Ao documentar requisitos, considere o comportamento atual do sistema. Documente se a alteração substitui, complementa ou é independente de funcionalidades existentes.

5. **O arquivo gerado é imutável para as etapas seguintes.** Após aprovação, o PRD não deve ser alterado sem um novo versionamento no histórico de revisões.

---

## Output esperado

```
.planning/
└── prd_{alteracao}.md   ← gerado nesta etapa
```

O nome `{alteracao}` deve ser em snake_case, descritivo e curto. Exemplos:
- `prd_autenticacao_oauth.md`
- `prd_exportacao_relatorios.md`
- `prd_correcao_calculo_frete.md`

---

## Transição para próxima etapa

Após gerar o PRD, informe ao usuário:

> "PRD gerado em `.planning/prd_{alteracao}.md`. Revise e, quando estiver satisfeito, me diga para avançar para o **SPEC** — onde vamos mapear os arquivos, tecnologias e plano de testes."
