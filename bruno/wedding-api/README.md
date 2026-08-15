# Wedding API — coleção Bruno

## Antes de começar

Inicie a API com um administrador configurado:

```bash
ADMIN_EMAIL=admin@example.com \
ADMIN_PASSWORD='uma-senha-forte' \
go run ./cmd
```

No Bruno, selecione o ambiente `local` e ajuste `adminEmail` e `adminPassword` em `environments/local.bru`.

O backend autentica por cookie `wedding_session`. O Bruno deve manter o cookie recebido nas respostas. Não é necessário preencher manualmente o header `Authorization`.

## Fluxo recomendado

1. Execute `Auth / Admin login`.
2. Execute `Admin / Create family` e copie o `id` retornado para `familyId`.
3. Execute `Admin / Create family access link`.
4. Copie o valor depois de `token=` da URL retornada para `familyToken`.
5. Execute `Auth / Exchange family token`. O cookie agora representa a família.
6. Execute as operações de família em `Family`, `Products` e `Presences`.

Para voltar ao admin, execute novamente `Auth / Admin login`. Para testar duas sessões simultâneas, use ambientes ou cookie jars separados no Bruno.

## Permissões

| Grupo | Admin | Família |
|---|---:|---:|
| `/api/auth/admin/login` | público | público |
| `/api/auth/exchange` | público | público |
| CRUD de famílias e links | sim | não |
| Visualizar produtos | sim | sim |
| Criar/editar/excluir produtos | sim | não |
| Reservar produto | sim | sim |
| Visualizar presenças | todas | somente a própria família |
| Criar/editar/excluir presenças | sim | não |
| Confirmar/cancelar presença | sim | somente a própria família |
| `/api/family` | não se aplica | própria família |

## Observações

- O link de família não expira automaticamente, mas pode ser revogado pelo admin.
- Revogar o link também revoga as sessões existentes daquela família.
- A relação do banco é `Family 1:N ConfirmationPresence`: cada pessoa pertence a uma família, e uma família pode ter várias pessoas.
- Como o login é por família, `reserved_by` é informado no corpo da requisição de reserva.
