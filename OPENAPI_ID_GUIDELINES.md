# OpenAPI Resource ID Rules

## Boundary

- Generic models keep their database primary key for relations and transactions.
- A channel OpenAPI exposes only `display_id` for a mock-owned resource. It never
  exposes the database primary key.
- An OpenAPI service resolves every incoming mock-owned resource ID by that
  channel's `display_id` before it calls its usecase. Usecases and repositories
  operate on the resolved database primary key.
- Merchant, member, and caller-supplied idempotency IDs are not mock resource
  IDs. They remain channel DTO fields and are not resolved as `display_id`.

## Channel Formats

| Channel | Resource | Display ID format |
| --- | --- | --- |
| PhotonPay | card holder | `CH` followed by decimal digits |
| PhotonPay | card | `XR` followed by decimal digits |
| PhotonPay | other mock-owned resources | opaque string; do not parse it as an integer |
| Paynda | all mock-owned resources | decimal digits encoded as a JSON string |
| Slash | all mock-owned resources | PostgreSQL UUIDv7 string |

The PhotonPay prefixes are established by the Marxo SDK fixtures. Paynda values
are passed and stored as strings even when a third-party response represents a
related ID as a JSON number. Slash SDK resource IDs are opaque strings; Marxo
does not parse them as integers.

## Required Mapping

- PhotonPay: resolve `cardholderId`, `cardId`, `transactionId`, and
  `originTransactionId` when those fields refer to resources created by this
  mock. Return the corresponding display IDs in all response DTOs.
- Paynda: resolve path and body resource IDs including `cardholderId`,
  `cardBinId`, `cardId`, and `transactionId`; return decimal display IDs in
  responses.
- Slash: resolve `:id`, `cardProductId`, `filter:cardId`, and
  `filter:providerAuthorizationId` when they name mock resources; return UUIDv7
  display IDs in cards, transactions, and card products.
