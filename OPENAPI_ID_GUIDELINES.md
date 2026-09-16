# OpenAPI Resource ID Rules

## Boundary

- Generic models keep their auto-incrementing database primary key for relations
  and transactions. They do not persist a `display_id` column.
- A channel OpenAPI derives the external ID of a mock-owned resource directly
  from its primary key through the channel's `service/id.go` formatter. The
  inverse parser runs in the service before it calls a usecase.
- Merchant, member, and caller-supplied idempotency IDs are not mock resource
  IDs and must not be passed through a channel ID parser.

## Channel Formats

| Channel | Resource | External ID format |
| --- | --- | --- |
| PhotonPay | all mock-owned resources | decimal digits encoded as a string |
| Paynda | all mock-owned resources | decimal digits encoded as a JSON string |
| Slash | all mock-owned resources | UUID-shaped string derived by the Slash formatter |

Paynda values are passed as strings even when a third-party response represents
a related ID as a JSON number. Do not introduce prefixes or random identifiers
unless the current channel formatter and verified downstream call sites require
them.

## Required Mapping

- PhotonPay: parse and format `cardholderId`, `cardId`, `transactionId`, and
  `originTransactionId` with the PhotonPay formatter when those fields refer to
  resources created by this mock.
- Paynda: parse and format path and body resource IDs including `cardholderId`,
  `cardBinId`, `cardId`, and `transactionId` with the Paynda formatter.
- Slash: parse and format `:id`, `cardProductId`, `filter:cardId`, and
  `filter:providerAuthorizationId` with the Slash formatter.
