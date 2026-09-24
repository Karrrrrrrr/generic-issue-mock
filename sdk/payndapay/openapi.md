# PayndaPay OpenAPI

**版本：** V1.6
**发布日期：** 2025-06-11

## 修订记录

| 版本   | 修订日期     | 修订说明                                                                                                                              |
| :----- | :----------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| V0.9   | 2025-03-30   | 草案版本 0.9                                                                                                                          |
| V1.0   | 2025-03-31   | 添加交易类型                                                                                                                          |
| V1.1   | 2025-04-15   | 添加卡 Bin 选择指南                                                                                                                   |
| V1.2   | 2025-04-21   | 卡 Bin 响应增加 `type` 和 `cardClass` 字段。                                                                                            |
| V1.3   | 2025-05-08   | 更新 `创建持卡人` 和 `创建资金账户` 的响应。调整章节顺序。添加 `卡片交易控制`。                                                              |
| V1.4   | 2025-05-09   | 卡片交易和 webhook 增加字段 `currency`, `transactionAmountInOriginalCurrency`, `originalCurrencyCode`                                     |
| V1.5   | 2025-05-14   | 为持卡人添加无限余额选项 (`unlimitedBalance`)。                                                                                           |
| V1.6   | 2025-06-11   | 卡片余额增加可用金额。为卡片添加 `creditLimitType` 属性以指示是共享额度还是独立额度。默认创建的卡片都是共享信用额度卡。详情请参见 `信用额度类型` 章节。获取 `CardBin` API 增加过滤参数 `creditLimitType`，默认返回 shared-limit 类型的卡 Bin。为 independent-limit 卡添加 `卡片充值/提现` API。请注意，原有的 `更新卡片额度` API 仅支持 shared-limit 卡。 |

---

# 基本信息

**端点：** https://vcc-openapi-prod.payndapay.com

# 生产环境上线

1.  Payndapay 在 PROD 环境中为客户注册商户、appId 和 appSecret。
2.  Payndapay 将向客户提供一个账户和密码。客户可以到我们的网站获取 appId 和 appSecret。
3.  客户提供 PROD 环境的 webhook 端点。

# 术语定义

*   **商户 (Merchant):** 您的公司账户。
*   **商户钱包 (Merchant Wallet):** 商户尚未分配到资金账户的资金。
*   **资金账户 (Balance Account):** 您公司名下的多个账户。每个资金账户的余额是相互隔离的。您可以从商户钱包向资金账户转账，或从资金账户向商户钱包转账。
*   **资金账户钱包 (Balance Account Wallet):** 每种货币的余额。创建新的资金账户后，您应设置资金账户的可用金额。
*   **持卡人 (Cardholder):** 一个资金账户下有多个持卡人。所有持卡人将共享该资金账户的余额。您可以为每个持卡人设置余额，但他们实际可用的消费金额受资金账户余额的限制，因为所有持卡人共享资金账户的余额。
*   **持卡人钱包 (Cardholder Wallet):** 持卡人每种货币的余额。创建新的持卡人后，您应设置持卡人的可用金额。
*   **卡片 (Card):** 信用卡。所有卡片都属于一个持卡人。
*   **卡片余额 (Card Balance):** 卡片的信用额度。

# API 认证

在调用 API 之前，请提供您的服务器 IP 列表。本服务设有白名单防火墙。调用 API 时，请确保正确配置以下请求头 (headers)：

| 请求头     | 值                         | 备注                                                                                                                               |
| :--------- | :------------------------- | :--------------------------------------------------------------------------------------------------------------------------------- |
| appId      |                            | 您的 appId                                                                                                                         |
| timestamp  |                            | 当前的 Unix 时间戳（秒）                                                                                                             |
| nonce      |                            | 每次请求生成一个唯一的 UUID。禁止在多个请求中重复使用相同的 UUID。这确保了每个请求的唯一性，并有助于防止重放攻击或重复提交。                           |
| sign       |                            | 计算签名的方法：md5(appId + appSecret + timestamp + nonce + path)。确保结果为小写。                                                              |

**签名示例：**

1.  appId=appId_yp0A2DPn1RxK
2.  appSecret=xxx
3.  timestamp=1718783677
4.  nonce=bla8ceafdble4cef8ec00faeacf75ac1
5.  path=/openapi/merchant/wallets
6.  sign=md5(appId_yp0A2DPn1RxKxxx1718783677bla8ceafdble4cef8ec00faeacf75ac1/openapi/merchant/wallets)

**Webhook 认证**

当接收 webhook 消息时，也会使用相同的方法对消息进行签名。将计算出的签名与 webhook 消息中提供的 sign 进行比较，以验证其真实性。

# 交易类型

| 交易类型                                  | 定义                                                                                             |
| :---------------------------------------- | :----------------------------------------------------------------------------------------------- |
| transaction.authentication.approved       | 卡片授权尝试已获批准。当使用卡片且供应商收到支付卡片详情后，商户授权了该交易时发生。                                 |
| transaction.authentication.declined       | 卡片授权尝试被拒绝。                                                                                   |
| transaction.authentication.settled        | 先前批准的交易已成功结算。                                                                               |
| transaction.authentication.reversal.pending | 先前批准的交易在结算完成前正被供应商作废。                                                              |
| transaction.authentication.reversal.settled | 先前批准的交易在结算过程之前已被供应商作废。                                                             |
| transaction.authentication.reversal.expired | 先前批准的交易由于 7 天后未收到结算或撤销而已过期。                                                              |
| transaction.refund.approved               | 尝试向卡片退款已发生并已获批准，但尚未完成。                                                                     |
| transaction.refund.settled                | 尝试向卡片退款已发生并已成功批准和结算。                                                                       |
| transaction.refund.declined               | 尝试向卡片退款已发生但被拒绝。                                                                             |
| transaction.refund.reversal               | 退款批准后，退款已被作废并撤销。                                                                      |

# 卡 Bin 选择指南

**对于美国境内消费，选择以下类型：**

*   Commercial Prepaid
*   Commercial Prepaid TE

**对于非美国境内消费，选择以下类型：**

*   GVCC200BPS

# 信用额度类型

我们提供两种类型的卡片：**共享额度 (shared-limit)** 卡和**独立额度 (independent-limit)** 卡。默认创建的卡片都是共享信用额度卡。这取决于您使用的卡 Bin。参见 [获取 CardBin API](#)。

## 共享额度卡 (Shared-limit card)

共享额度卡是一种虚拟卡，使用资金账户的总余额作为其额度，并支持消费额度控制。无需向卡片内转账——交易金额直接从资金账户中扣除。

同一资金账户下的所有共享额度卡共享该资金账户的总余额。一旦资金账户余额耗尽，所有共享额度卡都无法使用。

## 独立额度卡 (Independent-limit card)

独立额度卡是一种支持转入资金的虚拟卡。它支持充值和提现——卡片余额可以转出或在卡片删除时提取。

充值金额将从资金账户中扣除，任何提现金额将返回到资金账户。

# 商户钱包 (Merchant Wallet)

## 查询商户钱包

**基本信息**

**路径：** /openapi/merchant/wallets
**方法：** GET

**描述：**
查询商户不同货币的实时金额。

**参数**

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应 (Response)**

| 名称                | 类型      | 必需  | 默认值 | 备注             | 其他             |
| :------------------ | :-------- | :---- | :----- | :--------------- | :--------------- |
| code                | number    | N     |        |                  | format: int64    |
| message             | string    | N     |        |                  | format: string   |
| data                | object    | N     |        |                  |                  |
| ├─ pages            | number    | N     |        | 总页数           | format: int64    |
| ├─ records          | object [] | N     |        |                  | item Type: object|
| │  ├─ id            | string    | N     |        | 无注释           | format: string   |
| │  ├─ createTime    | string    | N     |        | 无注释           | format: string   |
| │  ├─ updateTime    | string    | N     |        | 无注释           | format: string   |
| │  ├─ merchantId    | string    | N     |        | 无注释           | format: string   |
| │  ├─ currency      | string    | N     |        | 无注释           | format: string   |
| │  ├─ amount        | string    | N     |        | 无注释           | format: string   |
| ├─ total            | number    | N     |        |                  | format: int64    |
| ├─ size             | number    | N     |        |                  | format: int64    |
| ├─ current          | number    | N     |        |                  | format: int64    |
| ├─ orders           | object [] | N     |        |                  | item Type: object|
| ├─ column           | string    | N     |        |                  | format: string   |
| ├─ asc              | boolean   | N     |        |                  |                  |
| ├─ optimizeCountSql | boolean   | N     |        |                  |                  |
| ├─ searchCount      | boolean   | N     |        |                  |                  |
| ├─ optimizeJoinOfCountSql | boolean | N | | {@link #optimizeJoinOfCountSql()} | |
| ├─ maxLimit         | number    | N     |        |                  | format: int64    |
| ├─ countId          | string    | N     |        | countId          | format: string   |
| ├─ success          | boolean   | N     |        |                  |                  |

---

# 资金账户 (Balance Account)

## 资金账户详情

**基本信息**
**路径：** /openapi/balanceAccounts/{id}
**方法：** GET

**描述：**
资金账户详情

**参数**
**路径参数 (Path Parameters)**

| 名称 | 示例 | 备注            |
| :--- | :--- | :-------------- |
| id   |      | 资金账户 ID       |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注       | 其他           |
| :------------------ | :------ | :---- | :----- | :--------- | :------------- |
| code                | number  | N     |        |            | format: int64  |
| message             | string  | N     |        |            | format: string |
| data                | object  | N     |        |            |                |
| ├─ id               | string  | N     |        | 无注释     | format: string |
| ├─ createTime       | string  | N     |        | 无注释     | format: string |
| ├─ updateTime       | string  | N     |        | 无注释     | format: string |
| ├─ name             | string  | N     |        | 名称       | format: string |
| ├─ merchantId       | string  | N     |        | 无注释     | format: string |
| success             | boolean | N     |        |            |                |

## 创建资金账户

**基本信息**
**路径：** /openapi/balanceAccounts
**方法：** POST
**描述：**
在商户下创建一个资金账户

**参数**

**请求体 (Body)**

| 名称 | 类型   | 必需  | 默认值 | 备注       | 其他           |
| :--- | :----- | :---- | :----- | :--------- | :------------- |
| name | string | N     |        | 无注释     | format: string |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他           |
| :------------------ | :------ | :---- | :----- | :------------- | :------------- |
| code                | number  | N     |        |                | format: int64  |
| message             | string  | N     |        |                | format: string |
| data                | object  | N     |        | (object)       |                |
| ├─ id               | string  | N     |        | 无注释         | format: string |
| ├─ createTime       | string  | N     |        | 无注释         | format: string |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string |
| ├─ name             | string  | N     |        | 名称           | format: string |
| ├─ merchantId       | string  | N     |        | 无注释         | format: string |
| success             | boolean | N     |        |                |                |

## 删除资金账户

**基本信息**
**路径：** /openapi/balanceAccounts/{id}
**方法：** DELETE
**描述：**
删除资金账户

**参数**

**路径参数 (Path Parameters)**

| 名称 | 示例 | 备注            |
| :--- | :--- | :-------------- |
| id   |      | 资金账户 ID       |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

## 查询资金账户

**基本信息**

**路径：** /openapi/balanceAccounts
**方法：** GET
**描述：**
查询商户下的所有资金账户信息

**参数**

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                | 类型      | 必需  | 默认值 | 备注             | 其他             |
| :------------------ | :-------- | :---- | :----- | :--------------- | :--------------- |
| code                | number    | N     |        |                  | format: int64    |
| message             | string    | N     |        |                  | format: string   |
| data                | object    | N     |        |                  |                  |
| ├─ pages            | number    | N     |        | 总页数           | format: int64    |
| ├─ records          | object [] | N     |        |                  | item Type: object|
| │  ├─ id            | string    | N     |        | 无注释           | format: string   |
| │  ├─ createTime    | string    | N     |        | 无注释           | format: string   |
| │  ├─ updateTime    | string    | N     |        | 无注释           | format: string   |
| │  ├─ name            | string    | N     |        | 名称             | format: string   |
| │  ├─ merchantId    | string    | N     |        | 无注释           | format: string   |
| ├─ total            | number    | N     |        |                  | format: int64    |
| ├─ size             | number    | N     |        |                  | format: int64    |
| ├─ current          | number    | N     |        |                  | format: int64    |
| ├─ orders           | object [] | N     |        |                  | item Type: object|
| ├─ column           | string    | N     |        |                  | format: string   |
| ├─ asc              | boolean   | N     |        |                  |                  |
| ├─ optimizeCountSql | boolean   | N     |        |                  |                  |
| ├─ searchCount      | boolean   | N     |        |                  |                  |
| ├─ optimizeJoinOfCountSql | boolean | N | | {@link #optimizeJoinOfCountSql()} | |
| ├─ maxLimit         | number    | N     |        |                  | format: int64    |
| ├─ countId          | string    | N     |        | countId          | format: string   |
| ├─ success          | boolean   | N     |        |                  |                  |

## 更新资金账户

**基本信息**
**路径：** /openapi/balanceAccounts/{id}
**方法：** PUT
**描述：**
更新资金账户

**参数**
**路径参数 (Path Parameters)**

| 名称 | 示例 | 备注            |
| :--- | :--- | :-------------- |
| id   |      | 资金账户 ID       |

**请求体 (Body)**

| 名称 | 类型   | 必需  | 默认值 | 备注       | 其他           |
| :--- | :----- | :---- | :----- | :--------- | :------------- |
| name | string | N     |        | 无注释     | format: string |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

---

# 资金账户钱包 (Balance Account Wallet)

## 查询资金账户钱包

**基本信息**
**路径：** /openapi/balanceAccounts/{id}/wallets
**方法：** GET

**描述：**
查询资金账户不同货币的实时金额

**参数**
**路径参数 (Path Parameters)**

| 名称 | 示例 | 备注            |
| :--- | :--- | :-------------- |
| id   |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                      | 类型      | 必需  | 默认值 | 备注                 | 其他             |
| :------------------------ | :-------- | :---- | :----- | :------------------- | :--------------- |
| code                      | number    | N     |        |                      | format: int64    |
| message                   | string    | N     |        |                      | format: string   |
| data                      | object    | N     |        |                      |                  |
| ├─ pages                  | number    | N     |        | 总页数               | format: int64    |
| ├─ records                | object [] | N     |        |                      | item Type: object|
| │  ├─ id                  | string    | N     |        | 无注释               | format: string   |
| │  ├─ createTime          | string    | N     |        | 无注释               | format: string   |
| │  ├─ updateTime          | string    | N     |        | 无注释               | format: string   |
| │  ├─ merchantId          | string    | N     |        | 无注释               | format: string   |
| │  ├─ balanceAccountId    | string    | N     |        | 无注释               | format: string   |
| │  ├─ currency            | string    | N     |        | 钱包的货币           | format: string   |
| │  ├─ amount              | string    | N     |        | 可用金额             | format: string   |
| │  ├─ frozenAmount        | string    | N     |        | 冻结金额             | format: string   |
| ├─ total                  | number    | N     |        |                      | format: int64    |
| ├─ size                   | number    | N     |        |                      | format: int64    |
| ├─ current                | number    | N     |        |                      | format: int64    |
| ├─ orders                 | object [] | N     |        |                      | item Type: object|
| ├─ column                 | string    | N     |        |                      | format: string   |
| ├─ asc                    | boolean   | N     |        |                      |                  |
| ├─ optimizeCountSql       | boolean   | N     |        |                      |                  |
| ├─ searchCount            | boolean   | N     |        |                      |                  |
| ├─ optimizeJoinOfCountSql | boolean   | N     |        | {@link #optimizeJoinOfCountSql()} | |
| ├─ maxLimit               | number    | N     |        |                      | format: int64    |
| ├─ countId                | string    | N     |        | countId              | format: string   |
| ├─ success                | boolean   | N     |        |                      |                  |

---

# 资金账户钱包转账 (Balance Account Wallet Transfer)

## 资金转账

**基本信息**

**路径：** /openapi/balanceAccountWalletTransfers
**方法：** POST

**描述：**
在商户和资金账户之间转移资金

**参数**

**请求体 (Body)**

| 名称               | 类型   | 必需  | 默认值 | 备注                                                                 | 其他                     |
| :----------------- | :----- | :---- | :----- | :------------------------------------------------------------------- | :----------------------- |
| balanceAccountId   | string | N     |        | 无注释                                                               | format: string           |
| type               | string | N     |        | IN: 从商户向资金账户转账 OUT: 从资金账户向商户转账 [Enum: IN(1), OUT(2)] | Enum: IN,OUT format: enum|
| currency           | string | N     |        | 无注释                                                               | format: string           |
| amount             | string | N     |        | 无注释                                                               | format: string           |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他                     |
| :------------------ | :------ | :---- | :----- | :------------- | :----------------------- |
| code                | number  | N     |        |                | format: int64            |
| message             | string  | N     |        |                | format: string           |
| data                | object  | N     |        |                |                          |
| ├─ id               | string  | N     |        | 无注释         | format: string           |
| ├─ createTime       | string  | N     |        |                |                          |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string           |
| ├─ merchantId       | string  | N     |        | 无注释         | format: string           |
| ├─ balanceAccountId | string  | N     |        | 无注释         | format: string           |
| ├─ type             | string  | N     |        | 无注释 [Enum: IN(1), OUT(2)] | Enum: IN,OUT format: enum|
| ├─ currency         | string  | N     |        | 无注释         | format: string           |
| ├─ amount           | string  | N     |        | 无注释         | format: string           |
| success             | boolean | N     |        |                |                          |

## 查询资金转账

**基本信息**
**路径：** /openapi/balanceAccountWalletTransfers
**方法：** GET
**描述：**
您可以使用此 API 查询商户和资金账户之间资金转账的详细信息。

**参数**

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                | 类型      | 必需  | 默认值 | 备注             | 其他                     |
| :------------------ | :-------- | :---- | :----- | :--------------- | :----------------------- |
| code                | number    | N     |        |                  | format: int64            |
| message             | string    | N     |        |                  | format: string           |
| data                | object    | N     |        |                  |                          |
| ├─ pages            | number    | N     |        | 总页数           | format: int64            |
| ├─ records          | object [] | N     |        |                  | item Type: object        |
| │  ├─ id            | string    | N     |        | 无注释           | format: string           |
| │  ├─ createTime    | string    | N     |        | 无注释           | format: string           |
| │  ├─ updateTime    | string    | N     |        | 无注释           | format: string           |
| │  ├─ merchantId    | string    | N     |        | 无注释           | format: string           |
| │  ├─ balanceAccountId | string | N     |        |                  |                          |
| │  ├─ type          | string    | N     |        | 无注释 [Enum: IN(1), OUT(2)] | Enum: IN,OUT format: enum|
| │  ├─ currency      | string    | N     |        | 无注释           | format: string           |
| │  ├─ amount        | string    | N     |        | 无注释           | format: string           |
| ├─ total            | number    | N     |        |                  | format: int64            |
| ├─ size             | number    | N     |        |                  | format: int64            |
| ├─ current          | number    | N     |        |                  | format: int64            |
| ├─ orders           | object [] | N     |        |                  | item Type: object        |
| ├─ column           | string    | N     |        |                  | format: string           |
| ├─ asc              | boolean   | N     |        |                  |                          |
| ├─ optimizeCountSql | boolean   | N     |        |                  |                          |
| ├─ searchCount      | boolean   | N     |        |                  |                          |
| ├─ optimizeJoinOfCountSql | boolean | N | | @link #optimizeJoinOfCountSql() | format: int64 |
| ├─ maxLimit         | number    | N     |        |                  |                          |
| ├─ countId          | string    | N     |        |                  | format: string           |
| ├─ success          | boolean   | N     |        |                  |                          |

---

# 持卡人 (Cardholder)

## 持卡人详情

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
**方法：** GET

**描述：**
持卡人详情

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 持卡人 ID        |

**响应**

| 名称                    | 类型    | 必需  | 默认值 | 备注                 | 其他           |
| :---------------------- | :------ | :---- | :----- | :------------------- | :------------- |
| code                    | number  | N     |        |                      | format: int64  |
| message                 | string  | N     |        |                      | format: string |
| data                    | object  | N     |        |                      |                |
| ├─ id                   | string  | N     |        | 无注释               | format: string |
| ├─ createTime           | string  | N     |        | 无注释               | format: string |
| ├─ updateTime           | string  | N     |        | 无注释               | format: string |
| ├─ merchantId           | string  | N     |        | 无注释               | format: string |
| ├─ balanceAccountId     | string  | N     |        | 无注释               | format: string |
| ├─ firstName            | string  | N     |        | 无注释               | format: string |
| ├─ lastName             | string  | N     |        | 无注释               | format: string |
| ├─ mobilePrefix         | string  | N     |        | 无注释               | format: string |
| ├─ mobile               | string  | N     |        | 无注释               | format: string |
| ├─ email                | string  | N     |        | 无注释               | format: string |
| ├─ billingAddressLine1  | string  | N     |        | 无注释               | format: string |
| ├─ billingAddressLine2  | string  | N     |        | 无注释               | format: string |
| ├─ billingCity          | string  | N     |        | 无注释               | format: string |
| ├─ billingCountryCode   | string  | N     |        | 无注释               | format: string |
| ├─ billingPostalCode    | string  | N     |        | 无注释               | format: string |
| ├─ billingState         | string  | N     |        | 无注释               | format: string |
| ├─ unlimitedBalance     | boolean | N     | false  | 无限余额选项。true: 无限。false: 有限，使用持卡人钱包。 | format: boolean|
| success                 | boolean | N     |        |                      |                |

## 创建持卡人

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders
**方法：** POST
**描述：**
创建持卡人

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**请求体 (Body)**

| 名称                | 类型    | 必需  | 默认值 | 备注                     | 其他           |
| :------------------ | :------ | :---- | :----- | :----------------------- | :------------- |
| firstName           | string  | N     |        | 无注释                   | format: string |
| lastName            | string  | N     |        | 无注释                   | format: string |
| mobilePrefix        | string  | N     |        | 国际电话区号，例如 86。      | format: string |
| mobile              | string  | N     |        | 手机号码。                 | format: string |
| email               | string  | N     |        | 邮箱地址。                 | format: string |
| billingAddressLine1 | string  | N     |        | 账单地址行 1              | format: string |
| billingAddressLine2 | string  | N     |        | 账单地址行 2              | format: string |
| billingCity         | string  | N     |        | 账单城市                 | format: string |
| billingCountryCode  | string  | N     |        | 账单国家代码。             | format: string |
| billingPostalCode   | string  | N     |        | 账单邮政编码。             | format: string |
| billingState        | string  | N     |        | 账单州或省。               | format: string |
| unlimitedBalance    | boolean | N     | false  | 无限余额选项。true: 无限。false: 有限，使用持卡人钱包。 | format: boolean|

**响应**

| 名称                    | 类型    | 必需  | 默认值 | 备注                 | 其他           |
| :---------------------- | :------ | :---- | :----- | :------------------- | :------------- |
| code                    | number  | N     |        |                      | format: int64  |
| message                 | string  | N     |        |                      | format: string |
| data                    | object  | N     |        | (object)             |                |
| ├─ id                   | string  | N     |        | 无注释               | format: string |
| ├─ createTime           | string  | N     |        | 无注释               | format: string |
| ├─ updateTime           | string  | N     |        | 无注释               | format: string |
| ├─ merchantId           | string  | N     |        | 无注释               | format: string |
| ├─ balanceAccountId     | string  | N     |        | 无注释               | format: string |
| ├─ firstName            | string  | N     |        | 无注释               | format: string |
| ├─ lastName             | string  | N     |        | 无注释               | format: string |
| ├─ mobilePrefix         | string  | N     |        | 无注释               | format: string |
| ├─ mobile               | string  | N     |        | 无注释               | format: string |
| ├─ email                | string  | N     |        | 无注释               | format: string |
| ├─ billingAddressLine1  | string  | N     |        | 无注释               | format: string |
| ├─ billingAddressLine2  | string  | N     |        | 无注释               | format: string |
| ├─ billingCity          | string  | N     |        | 无注释               | format: string |
| ├─ billingCountryCode   | string  | N     |        | 无注释               | format: string |
| ├─ billingPostalCode    | string  | N     |        | 无注释               | format: string |
| ├─ billingState         | string  | N     |        | 无注释               | format: string |
| ├─ unlimitedBalance     | boolean | N     | false  | 无限余额选项。true: 无限。false: 有限，使用持卡人钱包。 | format: boolean|

## 删除持卡人

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
**方法：** DELETE

**描述：**
删除持卡人

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 持卡人 ID        |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

## 查询持卡人

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders
**方法：** GET

**描述：**
查询持卡人

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                    | 类型      | 必需  | 默认值 | 备注                 | 其他           |
| :---------------------- | :-------- | :---- | :----- | :------------------- | :------------- |
| code                    | number    | N     |        |                      | format: int64  |
| message                 | string    | N     |        |                      | format: string |
| data                    | object    | N     |        |                      |                |
| ├─ pages                | number    | N     |        | 总页数               | format: int64  |
| ├─ records              | object [] | N     |        |                      | item Type: object|
| │  ├─ id                | string    | N     |        | 无注释               | format: string |
| │  ├─ createTime        | string    | N     |        | 无注释               | format: string |
| │  ├─ updateTime        | string    | N     |        | 无注释               | format: string |
| │  ├─ merchantId        | string    | N     |        | 无注释               | format: string |
| │  ├─ balanceAccountId  | string    | N     |        | 无注释               | format: string |
| │  ├─ firstName         | string    | N     |        | 无注释               | format: string |
| │  ├─ lastName          | string    | N     |        | 无注释               | format: string |
| │  ├─ mobilePrefix      | string    | N     |        | 无注释               | format: string |
| │  ├─ mobile            | string    | N     |        | 无注释               | format: string |
| │  ├─ email             | string    | N     |        | 无注释               | format: string |
| │  ├─ billingAddressLine1 | string  | N     |        | 无注释               | format: string |
| │  ├─ billingAddressLine2 | string  | N     |        | 无注释               | format: string |
| │  ├─ billingCity       | string    | N     |        | 无注释               | format: string |
| │  ├─ billingCountryCode| string    | N     |        | 无注释               | format: string |
| │  ├─ billingPostalCode | string    | N     |        | 无注释               | format: string |
| │  ├─ billingState      | string    | N     |        | 无注释               | format: string |
| │  ├─ unlimitedBalance  | boolean   | N     | false  | 无限余额选项。true: 无限。false: 有限，使用持卡人钱包。 | format: boolean|
| ├─ total                | number    | N     |        |                      | format: int64  |
| ├─ size                 | number    | N     |        |                      | format: int64  |
| ├─ current              | number    | N     |        |                      | format: int64  |
| ├─ orders               | object [] | N     |        |                      | item Type: object|
| ├─ column               | string    | N     |        |                      | format: string |
| ├─ asc                  | boolean   | N     |        |                      |                |
| ├─ optimizeCountSql     | boolean   | N     |        |                      |                |
| ├─ searchCount          | boolean   | N     |        |                      |                |
| ├─ optimizeJoinOfCountSql | boolean | N     |        | {@link #optimizeJoinOfCountSql()} | format: int64 |
| ├─ countId              | string    | N     |        | countId              | format: string |
| ├─ success              | boolean   | N     |        |                      |                |

## 更新持卡人

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders/{id}
**方法：** PUT

**描述：**
更新持卡人

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 持卡人 ID        |

**请求体 (Body)**

| 名称                | 类型    | 必需  | 默认值 | 备注                     | 其他           |
| :------------------ | :------ | :---- | :----- | :----------------------- | :------------- |
| firstName           | string  | N     |        | 无注释                   | format: string |
| lastName            | string  | N     |        | 无注释                   | format: string |
| mobilePrefix        | string  | N     |        | 国际电话区号，例如 86。      | format: string |
| mobile              | string  | N     |        | 手机号码                 | format: string |
| email               | string  | N     |        | 邮箱地址。                 | format: string |
| billingAddressLine1 | string  | N     |        | 账单地址行 1              | format: string |
| billingAddressLine2 | string  | N     |        | 账单地址行 2              | format: string |
| billingCity         | string  | N     |        | 账单城市                 | format: string |
| billingCountryCode  | string  | N     |        | 账单国家代码。             | format: string |
| billingPostalCode   | string  | N     |        | 账单邮政编码。             | format: string |
| billingState        | string  | N     |        | 账单州或省。               | format: string |
| unlimitedBalance    | boolean | N     |        | 无限余额选项。true: 无限。false: 有限，使用持卡人钱包。 | format: boolean|

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

---

# 持卡人钱包 (Cardholder Wallet)

## 查询持卡人钱包

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholders/{cardholderId}/wallets
**方法：** GET

**描述：**
查询持卡人不同货币的实时金额

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| cardholderId     |      | 持卡人 ID        |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                      | 类型      | 必需  | 默认值 | 备注                 | 其他             |
| :------------------------ | :-------- | :---- | :----- | :------------------- | :--------------- |
| code                      | number    | N     |        |                      | format: int64    |
| message                   | string    | N     |        |                      | format: string   |
| data                      | object    | N     |        |                      |                  |
| ├─ pages                  | number    | N     |        | 总页数               | format: int64    |
| ├─ records                | object [] | N     |        |                      | item Type: object|
| │  ├─ id                  | string    | N     |        | 无注释               | format: string   |
| │  ├─ createTime          | string    | N     |        | 无注释               | format: string   |
| │  ├─ updateTime          | string    | N     |        | 无注释               | format: string   |
| │  ├─ merchantId          | string    | N     |        | 无注释               | format: string   |
| │  ├─ balanceAccountId    | string    | N     |        | 无注释               | format: string   |
| │  ├─ cardholderId        | string    | N     |        | 无注释               | format: string   |
| │  ├─ currency            | string    | N     |        | 钱包的货币           | format: string   |
| │  ├─ amount              | string    | N     |        | 可用金额             | format: string   |
| │  ├─ frozenAmount        | string    | N     |        | 冻结金额             | format: string   |
| ├─ total                  | number    | N     |        |                      | format: int64    |
| ├─ size                   | number    | N     |        |                      | format: int64    |
| ├─ current                | number    | N     |        |                      | format: int64    |
| ├─ orders                 | object [] | N     |        |                      | item Type: object|
| ├─ column                 | string    | N     |        |                      | format: string   |
| ├─ asc                    | boolean   | N     | |                      |                  |
| ├─ optimizeCountSql       | boolean   | N     |        |                      |                  |
| ├─ searchCount            | boolean   | N     |        |                      |                  |
| ├─ optimizeJoinOfCountSql | boolean   | N     |        | @link #optimizeJoinOfCountSql() | |
| ├─ maxLimit               | number    | N     |        |                      | format: int64    |
| ├─ countId                | string    | N     |        | countId              | format: string   |
| ├─ success                | boolean   | N     |        |                      |                  |

## 查询持卡人钱包更新历史

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholderWalletUpdates
**方法：** GET

**描述：**
查询持卡人钱包更新历史

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                      | 类型      | 必需  | 默认值 | 备注                 | 其他                     |
| :------------------------ | :-------- | :---- | :----- | :------------------- | :----------------------- |
| code                      | number    | N     |        |                      | format: int64            |
| message                   | string    | N     |        |                      | format: string           |
| data                      | object    | N     |        |                      |                          |
| ├─ pages                  | number    | N     |        | 总页数               | format: int64            |
| ├─ records                | object [] | N     |        |                      | item Type: object        |
| │  ├─ id                  | string    | N     |        | 无注释               | format: string           |
| │  ├─ createTime          | string    | N     |        | 无注释               | format: string           |
| │  ├─ updateTime          | string    | N     |        | 无注释               | format: string           |
| │  ├─ merchantId          | string    | N     |        | 无注释               | format: string           |
| │  ├─ balanceAccountId    | string    | N     |        | 无注释               | format: string           |
| │  ├─ cardholderId        | string    | N     |        | 无注释               | format: string           |
| │  ├─ type                | string    | N     |        | 无注释 [Enum: INC(1), DEC(2)] | Enum: INC,DEC format: enum|
| │  ├─ currency            | string    | N     |        | 无注释               | format: string           |
| │  ├─ amount              | string    | N     |        | 无注释               | format: string           |
| ├─ total                  | number    | N     |        |                      | format: int64            |
| ├─ size                   | number    | N     |        |                      | format: int64            |
| ├─ current                | number    | N     |        |                      | format: int64            |
| ├─ orders                 | object [] | N     |        |                      | item Type: object        |
| ├─ column                 | string    | N     | |                      | format: string           |
| ├─ asc                    | boolean   | N     |        |                      |                          |
| ├─ optimizeCountSql       | boolean   | N     |        |                      |                          |
| ├─ searchCount            | boolean   | N     |        |                      |                          |
| ├─ optimizeJoinOfCountSql | boolean   | N     |        | @link #optimizeJoinOfCountSql() | |
| ├─ maxLimit               | number    | N     |        |                      | format: int64            |
| ├─ countId                | string    | N     |        | countId              | format: string           |
| ├─ success                | boolean   | N     |        |                      |                          |

## 更新持卡人钱包金额

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardholderWalletUpdates
**方法：** POST
**描述：**
更新持卡人钱包金额

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**请求体 (Body)**

| 名称          | 类型   | 必需  | 默认值 | 备注                                       | 其他                     |
| :------------ | :----- | :---- | :----- | :----------------------------------------- | :----------------------- |
| cardholderId  | string | N     |        | 无注释                                     | format: string           |
| type          | string | N     |        | INC: 增加 DEC: 减少 [Enum: INC(1), DEC(2)] | Enum: INC,DEC format: enum|
| currency      | string | N     |        | 无注释                                     | format: string           |
| amount        | string | N     |        | 增加/减少的金额                              | format: string           |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他                     |
| :------------------ | :------ | :---- | :----- | :------------- | :----------------------- |
| code                | number  | N     |        |                | format: int64            |
| message             | string  | N     |        |                | format: string           |
| data                | object  | N     |        |                |                          |
| ├─ id               | string  | N     |        | 无注释         | format: string           |
| ├─ createTime       | string  | N     |        | 无注释         | format: string           |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string           |
| ├─ merchantId       | string  | N     |        | 无注释         | format: string           |
| ├─ balanceAccountId | string  | N     | | 无注释         | format: string           |
| ├─ cardholderId     | string  | N     |        | 无注释         | format: string           |
| ├─ type             | string  | N     |        | 无注释 [Enum: INC(1), DEC(2)] | Enum: INC,DEC format: enum|
| ├─ currency         | string  | N     |        | 无注释         | format: string           |
| ├─ amount           | string  | N     |        | 无注释         | format: string           |
| success             | boolean | N     |        |                |                          |

---

# 卡 Bin (CardBin)

## 获取卡 Bin

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardBins
**方法：** GET
**描述：**
获取卡 Bin

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称            | 必需  | 默认值  | 示例                | 备注                                                |
| :-------------- | :---- | :------ | :------------------ | :-------------------------------------------------- |
| creditLimitType | N     | SHARED  | SHARED, INDEPENDENT | SHARED: 共享额度, INDEPENDENT: 独立额度。默认返回 shared-limit 类型。 |

**响应**

| 名称                | 类型      | 必需  | 默认值 | 备注                                                                 | 其他             |
| :------------------ | :-------- | :---- | :----- | :------------------------------------------------------------------- | :--------------- |
| code                | number    | N     |        |                                                                      | format: int64    |
| message             | string    | N     |        |                                                                      | format: string   |
| data                | object [] | N     |        |                                                                      | item Type: object|
| │  ├─ id            | string    | N     |        | 项目 id                                                              | format: string   |
| │  ├─ cardBin       | string    | N     |        | 无注释                                                               | format: string   |
| │  ├─ name          | string    | N     |        | 无注释                                                               | format: string   |
| │  ├─ type          | string    | N     |        | CARD_CLASS: 此选择项是一个卡片类别。这将确保从此类别中选择一个 BIN。                |                  |
| │                  |           |       |        | CARD_BIN: 此选择项是一个特定的卡 bin。                                  |                  |
| │  ├─ cardClass     | string    |       |        | 卡片类别                                                               |                  |
| │  ├─ supportCurrencies | string [] | N  |        | 支持的货币                                                             | item Type: string|
| success             | boolean   | N     |        |                                                                      |                  |

---

# 卡片余额 (Card Balance)

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/balance
**方法：** GET

**描述：**
卡片余额

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他           |
| :------------------ | :------ | :---- | :----- | :------------- | :------------- |
| code                | number  | N     |        |                | format: int64  |
| message             | string  | N     |        |                | format: string |
| data                | object  | N     |        |                |                |
| ├─ id               | string  | N     |        | 无注释         | format: string |
| ├─ createTime       | string  | N     |        | 无注释         | format: string |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string |
| ├─ amountUsed       | string  | N     |        | 已结算金额       | format: string |
| ├─ amountFrozen     | string  | N     |        | 冻结金额         | format: string |
| ├─ availableAmount  | string  | N     |        | 可用金额         | format: string |
| ├─ amount           | string  | N     |        | 信用额度         | format: string |
| success             | boolean | N     |        |                |                |

---

# 卡片详情

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}
**方法：** GET

**描述：**
卡片详情

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                    | 类型    | 必需  | 默认值 | 备注                                                                                             | 其他                             |
| :---------------------- | :------ | :---- | :----- | :----------------------------------------------------------------------------------------------- | :------------------------------- |
| code                    | number  | N     |        |                                                                                                  | format: int64                    |
| message                 | string  | N     |        |                                                                                                  | format: string                   |
| data                    | object  | N     |        |                                                                                                  |                                  |
| ├─ id                   | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ createTime           | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ updateTime           | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ merchantId           | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ balanceAccountId     | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ cardholderId         | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ status               | string  | N     |        | 无注释 [Enum: DELETED(0), ACTIVE(1), FROZEN(2), EXPIRED(3), BLOCK(4), UNACTIVE(9), UNKNOWN(9999)] | Enum: DELETED,ACTIVE,FROZEN,EXPIRED,BLOCK,UNACTIVE,UNKNOWN format: enum |
| ├─ maskCardNo           | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ cardScheme           | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ cardBin              | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ currency             | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ firstName            | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ lastName             | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ mobilePrefix         | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ mobile               | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ email                | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingCountryCode   | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingAddressLine1  | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingAddressLine2  | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingCity          | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingPostalCode    | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ billingState         | string  | N     |        | 州代码                                                                                           | format: string                   |
| ├─ singleUse            | boolean | N     |        | 单次使用标志                                                                                       | format: boolean                  |
| ├─ creditLimitType      | enum    | N     |        | 信用额度类型。SHARED: 共享额度, INDEPENDENT: 独立额度。                                                 | format: string                   |
| success                 | boolean | N     |        |                                                                                                  |                                  |

## 创建卡片

**基本信息**

**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards
**方法：** POST

**描述：**
创建卡片

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**请求体 (Body)**

| 名称                      | 类型    | 必需  | 默认值  | 备注                                       | 其他           |
| :------------------------ | :------ | :---- | :------ | :----------------------------------------- | :------------- |
| cardholderId              | string  | Y     |         | 无注释                                     | format: string |
| currency                  | string  | Y     |         | 无注释                                     | format: string |
| amount                    | string  | Y     |         | 额度金额                                   | format: string |
| expirationDate            | string  | Y     |         | 过期日期。格式: MM/YY                        | format: string |
| cardBinId                 | string  | Y     |         | 卡 bin id                                  | format: string |
| singleUse                 | boolean | N     | false   | 单次使用卡                                   |                |
| transactionCountLimitTotal| number  | N     | 0       | 如果不是单次使用卡，可以限制交易总次数。如果为 0，则表示无限制。 |                |

**响应**

| 名称                    | 类型    | 必需  | 默认值 | 备注                                                                                             | 其他                             |
| :---------------------- | :------ | :---- | :----- | :----------------------------------------------------------------------------------------------- | :------------------------------- |
| code                    | number  | N     |        |                                                                                                  | format: int64                    |
| message                 | string  | N     |        |                                                                                                  | format: string                   |
| data                    | object  | N     |        |                                                                                                  |                                  |
| ├─ card                 | object  | N     |        |                                                                                                  |                                  |
| │  ├─ id                | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ createTime        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ updateTime        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ merchantId        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ balanceAccountId  | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardholderId      | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ status            | string  | N     |        | 无注释 [Enum: DELETED(0), ACTIVE(1), FROZEN(2), EXPIRED(3), BLOCK(4), UNACTIVE(9), UNKNOWN(9999)] | Enum: DELETED,ACTIVE,FROZEN,EXPIRED,BLOCK,UNACTIVE,UNKNOWN format: enum |
| │  ├─ maskCardNo        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardScheme        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardBin           | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ currency          | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ firstName         | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ lastName          | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ mobilePrefix      | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ mobile            | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ email             | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingCountryCode| string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingAddressLine1 | string | N  |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingAddressLine2 | string | N  |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingCity       | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingPostalCode | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingState      | string  | N     |        | 美国使用两个字母的州代码                                                                             | format: string                   |
| │  ├─ amountUsed        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ amount            | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ sensitiveInfo        | object  | N     |        | 无注释                                                                                           |                                  |
| │  ├─ id                | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ createTime        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ updateTime        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardId            | number  | N     |        | 无注释                                                                                           | format: int64                    |
| │  ├─ cvv               | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ expirationDate    | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardNo            | string  | N     |        | 无注释                                                                                           | format: string                   |
| ├─ balance              | object  | N     |        |                                                                                                  |                                  |
| │  ├─ id                | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ createTime        | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ updateTime        | string  | N     | | 无注释                                                                                           | format: string                   |
| │  ├─ amountUsed        | string  | N     |        | 结算金额                                                                                         | format: string                   |
| │  ├─ amountFrozen      | string  | N     |        | 冻结金额                                                                                         | format: string                   |
| │  ├─ amount            | string  | N     |        | 额度金额                                                                                         | format: string                   |
| success                 | boolean | N     |        |                                                                                                  |                                  |

## 获取卡片敏感信息

**基本信息**

**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/sensitiveInfo
**方法：** GET

**描述：**
获取卡片敏感信息

**参数**

**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他           |
| :------------------ | :------ | :---- | :----- | :------------- | :------------- |
| code                | number  | N     |        |                | format: int64  |
| message             | string  | N     |        |                | format: string |
| data                | object  | N     |        |                |                |
| ├─ id               | string  | N     |        | 无注释         | format: string |
| ├─ createTime       | string  | N     |        | 无注释         | format: string |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string |
| ├─ cardId           | number  | N     |        | 无注释         | format: int64  |
| ├─ cvv              | string  | N     |        | 无注释         | format: string |
| ├─ expirationDate   | string  | N     |        | 无注释         | format: string |
| ├─ cardNo           | string  | N     |        | 无注释         | format: string |
| success             | boolean | N     |        |                |                |

## 冻结卡片

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/frozen
**方法：** PATCH
**描述：**
冻结卡片

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

## 查询卡片

**基本信息**

**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards
**方法：** GET

**描述：**
查询卡片

**参数**

**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注        |
| :--------------- | :--- | :---------- |
| balanceAccountId |      | 无注释        |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                    | 类型      | 必需  | 默认值 | 备注                                                                                             | 其他                             |
| :---------------------- | :-------- | :---- | :----- | :----------------------------------------------------------------------------------------------- | :------------------------------- |
| code                    | number    | N     |        |                                                                                                  | format: int64                    |
| message                 | string    | N     |        |                                                                                                  | format: string                   |
| data                    | object    | N     |        |                                                                                                  |                                  |
| ├─ pages                | number    | N     |        | 总页数                                                                                           | format: int64                    |
| ├─ records              | object [] | N     |        |                                                                                                  | item Type: object                |
| │  ├─ id                | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ createTime        | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ updateTime        | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ merchantId        | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ balanceAccountId  | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardholderId      | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ status            | string    | N     |        | 无注释 [Enum: DELETED(0), ACTIVE(1), FROZEN(2), EXPIRED(3), BLOCK(4), UNACTIVE(9), UNKNOWN(9999)] | Enum: DELETED,ACTIVE,FROZEN,EXPIRED,BLOCK,UNACTIVE,UNKNOWN format: enum |
| │  ├─ maskCardNo        | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardScheme        | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ cardBin           | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ currency          | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ firstName         | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ lastName          | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ mobilePrefix      | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ mobile            | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ email             | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingCountryCode| string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingAddressLine1 | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingAddressLine2 | string  | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingCity       | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingPostalCode | string    | N     |        | 无注释                                                                                           | format: string                   |
| │  ├─ billingState      | string    | N     |        | 美国使用两个字母的州代码                                                                             | format: string                   |
| │  ├─ singleUse         | boolean   | N     |        | 单次使用标志                                                                                       | format: boolean                  |
| │  ├─ creditLimitType   | enum      | N     |        | 信用额度类型。SHARED: 共享额度, INDEPENDENT: 独立额度。                                                 | format: string                   |
| ├─ total                | number    | N     |        |                                                                                                  | format: int64                    |
| ├─ size                 | number    | N     |        |                                                                                                  | format: int64                    |
| ├─ current              | number    | N     |        |                                                                                                  | format: int64                    |
| ├─ orders               | object [] | N     |        |                                                                                                  | item Type: object                |
| ├─ column               | string    | N     |        |                                                                                                  | format: string                   |
| ├─ asc                  | boolean   | N     |        |                                                                                                  |                                  |
| ├─ optimizeCountSql     | boolean   | N     |        |                                                                                                  |                                  |
| ├─ searchCount          | boolean   | N     |        |                                                                                                  |                                  |
| ├─ optimizeJoinOfCountSql | boolean | N     |        | {@link #optimizeJoinOfCountSql()}                                                                | format: int64                    |
| ├─ maxLimit             | number    | N     |        |                                                                                                  |                                  |
| ├─ countId              | string    | N     |        | countId                                                                                          | format: string                   |
| ├─ success              | boolean   | N     |        |                                                                                                  |                                  |

## 解冻卡片

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/unfrozen
**方法：** PATCH

**描述：**
解冻卡片

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

## 释放卡片

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/status/release
**方法：** PATCH
**描述：**
释放卡片

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        | (object)     |                |
| success             | boolean | N     |        |              |                |

## 获取卡片交易控制

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/controls
**方法：** GET

**描述：**
获取卡片交易控制。每个项目代表一个特定时间段的限制。

**参数**

**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注                                                                                             | 其他                     |
| :------------------ | :------ | :---- | :----- | :---------------------------------------------------------------------------------------------- | :----------------------- |
| code                | number  | N     |        |                                                                                                 | format: int64            |
| message             | string  | N     |        |                                                                                                 | format: string           |
| data                | object []| N     |        |                                                                                                 |                          |
| ├─ period           | string  | Y     |        | DAY: 单日交易限额。MONTH: 月度交易限额。TOTAL: 累计交易限额。ONCE: 单笔交易限额。                         | Enum: DAY, MONTH, TOTAL, ONCE format: enum |
| ├─ transactionCount | number  | N     |        | 周期内的交易次数限制。当 period 为 ONCE 时，此字段不使用。当它小于或等于 0 时，表示没有限制。                 | format: int64            |
| ├─ amount           | string  | N     |        | 周期内的交易金额限制。当它小于或等于 0 时，表示没有限制。                                             | format: string           |
| success             | boolean | N     |        |                                                                                                 |                          |

## 更新卡片交易控制

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cards/{id}/controls
**方法：** PATCH
**描述：**
更新卡片交易控制。

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |
| id               |      | 卡片 ID         |

**请求体 (Body)**

| 名称                | 类型   | 必需  | 默认值 | 备注                                                                                             | 其他                     |
| :------------------ | :----- | :---- | :----- | :---------------------------------------------------------------------------------------------- | :----------------------- |
| period              | string | Y     |        | DAY: 单日交易限额。MONTH: 月度交易限额。TOTAL: 累计交易限额。ONCE: 单笔交易限额。                         |                          |
| transactionCount    | number | Y     |        | 周期内的交易次数限制。当 period 为 ONCE 时，此字段不使用。当它小于或等于 0 时，表示没有限制。                 | format: int64            |
| amount              | string | Y     |        | 周期内的交易金额限制。当它小于或等于 0 时，表示没有限制。                                             | format: string           |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注         | 其他           |
| :------------------ | :------ | :---- | :----- | :----------- | :------------- |
| code                | number  | N     |        |              | format: int64  |
| message             | string  | N     |        |              | format: string |
| data                | object  | N     |        |              |                |
| success             | boolean | N     |        |              |                |

---

# 卡片余额 (Card Balance)

## 查询更新历史

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardBalanceUpdates
**方法：** GET
**描述：**
查询更新历史

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称        | 示例  | 备注                   |
| :---------- | :---- | :--------------------- |
| current     | 1     | 当前页码，从 1 开始        |
| pageSize    | 10    | 每页大小                 |

**响应**

| 名称                | 类型      | 必需  | 默认值 | 备注             | 其他           |
| :------------------ | :-------- | :---- | :----- | :--------------- | :------------- |
| code                | number    | N     |        |                  | format: int64  |
| message             | string    | N     |        |                  | format: string |
| data                | object    | N     |        |                  |                |
| ├─ pages            | number    | N     |        | 总页数           | format: int64  |
| ├─ records          | object [] | N     |        |                  | item Type: object|
| │  ├─ id            | string    | N     |        | 无注释           | format: string |
| │  ├─ createTime    | string    | N     |        | 无注释           | format: string |
| │  ├─ updateTime    | string    | N     |        | 无注释           | format: string |
| │  ├─ merchantId    | string    | N     |        | 无注释           | format: string |
| │  ├─ balanceAccountId | string | N     |        | 无注释           | format: string |
| │  ├─ cardholderId  | string    | N     |        | 无注释           | format: string |
| │  ├─ cardId        | string    | N     |        | 无注释           | format: string |
| │  ├─ oldAmount     | string    | N     |        | 无注释           | format: string |
| │  ├─ newAmount     | string    | N     |        | 无注释           | format: string |
| ├─ total            | number    | N     |        |                  | format: int64  |
| ├─ size             | number    | N     |        |                  | format: int64  |
| ├─ current          | number    | N     |        |                  | format: int64  |
| ├─ orders           | object [] | N     |        |                  | item Type: object|
| ├─ column           | string    | N     |        |                  | format: string |
| ├─ asc              | boolean   | N     |        |                  |                |
| ├─ optimizeCountSql | boolean   | N     |        |                  |                |
| ├─ searchCount      | boolean   | N     |        |                  |                |
| ├─ optimizeJoinOfCountSql | boolean | N | | {@link #optimizeJoinOfCountSql()} | |
| ├─ maxLimit         | number    | N     |        |                  | format: int64  |
| ├─ countId          | string    | N     |        | countId          | format: string |
| ├─ success          | boolean   | N     |        |                  |                |

## 更新卡片额度

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardBalanceUpdates
**方法：** POST
**描述：**
更新共享额度卡的额度

**参数**

**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**请求体 (Body)**

| 名称      | 类型   | 必需  | 默认值 | 备注           | 其他           |
| :-------- | :----- | :---- | :----- | :------------- | :------------- |
| cardId    | string | N     |        | 无注释         | format: string |
| amount    | string | N     |        | 卡片额度金额     | format: string |

**响应**

| 名称                | 类型    | 必需  | 默认值 | 备注           | 其他           |
| :------------------ | :------ | :---- | :----- | :------------- | :------------- |
| code                | number  | N     |        |                | format: int64  |
| message             | string  | N     |        |                | format: string |
| data                | object  | N     |        |                |                |
| ├─ id               | string  | N     |        | 无注释         | format: string |
| ├─ createTime       | string  | N     |        | 无注释         | format: string |
| ├─ updateTime       | string  | N     |        | 无注释         | format: string |
| ├─ merchantId       | string  | N     |        | 无注释         | format: string |
| ├─ balanceAccountId | string  | N     |        | 无注释         | format: string |
| ├─ cardholderId     | string  | N     |        | 无注释         | format: string |
| ├─ cardId           | string  | N     |        | 无注释         | format: string |
| ├─ oldAmount        | string  | N     |        | 无注释         | format: string |
| ├─ newAmount        | string  | N     |        | 无注释         | format: string |
| success             | boolean | N     |        |                |                |

## 卡片充值/提现

**基本信息**
**路径：** /openapi/balanceAccounts/{balanceAccountId}/cardBalanceTransfers
**方法：** POST
**描述：**
为独立额度卡充值/提现。

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**请求体 (Body)**

| 名称   | 类型   | 必需  | 默认值 | 备注                 | 其他                     |
| :----- | :----- | :---- | :----- | :------------------- | :----------------------- |
| cardId | string | Y     |        | 无注释               | format: string           |
| amount | string | Y     |        | 充值/提现金额          | format: string           |
| type   | string | Y     |        | IN: 充值 OUT: 提现    | Enum: IN,OUT format: enum|

**响应**

| 名称      | 类型   | 必需  | 默认值 | 备注              | 其他                     |
| :-------- | :----- | :---- | :----- | :---------------- | :----------------------- |
| code      | number | N     |        |                   | format: int64            |
| message   | string | N     |        |                   | format: string           |
| data      | object | N     |        |                   |                          |
| ├─ cardId | string | Y     |        | 无注释            | format: string           |
| ├─ amount | string | Y     |        | 充值/提现金额       | format: string           |
| ├─ type   | string | Y     |        | IN: 充值 OUT: 提现 | Enum: IN,OUT format: enum|

---

# 卡片交易 (Card Transaction)

## 查询交易

**基本信息**

**路径：** /openapi/balanceAccounts/{balanceAccountId}/transactions
**方法：** GET

**描述：**
查询资金账户下的所有交易

**参数**
**路径参数 (Path Parameters)**

| 名称             | 示例 | 备注            |
| :--------------- | :--- | :-------------- |
| balanceAccountId |      | 资金账户 ID       |

**查询参数 (Query Parameters)**

| 名称                  | 示例 | 备注               |
| :-------------------- | :--- | :----------------- |
| current               | 1    | 当前页码，从 1 开始   |
| pageSize              | 10   | 每页大小             |
| balanceAccountId      |      | 过滤资金账户         |
| cardholderId          |      | 过滤持卡人           |
| cardId                |      | 过滤卡片             |
| transactionTimeStart  |      | 交易开始时间         |
| transactionTimeEnd    |      | 交易结束时间         |

**响应**

| 名称                                | 类型      | 必需  | 默认值 | 备注                         | 其他           |
| :---------------------------------- | :-------- | :---- | :----- | :--------------------------- | :------------- |
| code                                | number    | N     |        |                              | format: int64  |
| message                             | string    | N     |        |                              | format: string |
| data                                | object    | N     |        |                              |                |
| ├─ pages                            | number    | N     |        |                              |                |
| ├─ records                          | object [] | N     |        | 总页数                       | format: int64  |
| │  ├─ id                            | string    | N     |        | 无注释                       | format: string |
| │  ├─ createTime                    | string    | N     |        | 无注释                       | format: string |
| │  ├─ updateTime                    | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantId                    | number    | N     |        | 无注释                       | format: int64  |
| │  ├─ balanceAccountId              | number    | N     |        | 无注释                       | format: int64  |
| │  ├─ cardholderId                  | number    | N     |        | 无注释                       | format: int64  |
| │  ├─ cardId                        | number    | N     |        | 无注释                       | format: int64  |
| │  ├─ maskCardNo                    | string    | N     |        | 无注释                       | format: string |
| │  ├─ type                          | string    | N     |        | 无注释                       | format: string |
| │  ├─ approvalCode                  | string    | N     |        | 无注释                       | format: string |
| │  ├─ preAuthAmount                 | string    | N     |        | 无注释                       | format: string |
| │  ├─ postedAmount                  | string    | N     |        | 无注释                       | format: string |
| │  ├─ currency                      | string    | N     |        | 无注释                       | format: string |
| │  ├─ originalCurrencyCode          | string    | N     |        | 原始货币代码                   | format: string |
| │  ├─ transactionAmountInOriginalCurrency | string | N | | 原始货币中的交易金额             | format: string |
| │  ├─ reversalFlag                  | string    | N     |        | 无注释                       | format: string |
| │  ├─ transactionTime               | string    | N     |        | 无注释                       | format: string |
| │  ├─ authorizationTime             | string    | N     |        | 无注释                       | format: string |
| │  ├─ acquirerId                    | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantMcc                   | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantName                  | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantAddressAddressLine1   | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantAddressCity           | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantAddressState          | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantAddressCountry        | string    | N     |        | 无注释                       | format: string |
| │  ├─ merchantAddressZip            | string    | N     |        | 无注释                       | format: string |
| │  ├─ posAcceptorId                 | string    | N     |        | 无注释                       | format: string |
| │  ├─ posAcceptLocation             | string    | N     |        | 无注释                       | format: string |
| │  ├─ posEntryDescription           | string    | N     |        | 无注释                       | format: string |
| │  ├─ supplierTransactionId         | string    | N     |        | 无注释                       | format: string |
| ├─ total                            | number    | N     |        |                              | format: int64  |
| ├─ size                             | number    | N     |        |                              | format: int64  |
| ├─ current                          | number    | N     |        |                              | format: int64  |
| ├─ orders                           | object [] | N     |        |                              | item Type: object|
| ├─ column                           | string    | N     |        |                              | format: string |
| ├─ asc                              | boolean   | N     |        |                              |                |
| ├─ optimizeCountSql                 | boolean   | N     |        |                              |                |
| ├─ searchCount                      | boolean   | N     |        |                              |                |
| ├─ optimizeJoinOfCountSql           | boolean   | N     |        | {@link #optimizeJoinOfCountSql()} |                |
| ├─ maxLimit                         | number    | N     |        |                              | format: int64  |
| ├─ countId                          | string    | N     |        | counted                      | format: string |
| ├─ success                          | boolean   | N     |        |                              |                |

---

# Webhook

## Webhook

**基本信息**
**路径：** your-webhook-endpoint
**方法：** POST
**描述：**

**参数**
**请求头 (Headers)**

| 名称                          | 值                 | 必需  | 示例 | 备注                      |
| :---------------------------- | :----------------- | :---- | :--- | :------------------------ |
| X-VK-NOTIFICATION-CATEGORY    |                    | Y     |      | CARD_STATUS, CARD_TRANSACTION |

**请求体 (Body)**

| 名称                    | 类型    | 必需  | 默认值 | 备注           | 其他           |
| :---------------------- | :------ | :---- | :----- | :------------- | :------------- |
| category                | string  | Y     |        |                |                |
| cardStatusWebhook       | object  | N     |        |                |                |
| ├─ id                   | string  | N     |        | 无注释         | format: string |
| ├─ createTime           | string  | N     |        | 无注释         | format: string |
| ├─ updateTime           | string  | N     |        | 无注释         | format: string |
| ├─ merchantId           | number  | N     |        | 无注释         | format: int64  |
| ├─ balanceAccountId     | number  | N     |        | 无注释         | format: int64  |
| ├─ cardholderId         | number  | N     |        | 无注释         | format: int64  |
| ├─ cardId               | number  | N     |        | 无注释         | format: int64  |
| ├─ maskCardNo           | string  | N     |        | 无注释         | format: string |
| ├─ status               | string  | N     |        | DELETED, ACTIVE, FROZEN, EXPIRED, BLOCK, UNKNOWN | |
| cardTransactionWebhook  | object  | N     |        |                |                |
| ├─ id                   | string  | N     |        | 无注释         | format: string |
| ├─ createTime           | string  | N     |        | 无注释         | format: string |
| ├─ updateTime           | string  | N     |        | 无注释         | format: string |
| ├─ merchantId           | number  | N     |        | 无注释         | format: int64  |
| ├─ balanceAccountId     | number  | N     |        | 无注释         | format: int64  |
| ├─ cardholderId         | number  | N     |        | 无注释         | format: int64  |
| ├─ cardId               | number  | N     |        | 无注释         | format: int64  |
| ├─ maskCardNo           | string  | N     |        | 无注释         | format: string |
| ├─ type                 | string  | N     |        | 无注释         | format: string |
| ├─ approvalCode         | string  | N     |        | 无注释         | format: string |
| ├─ preAuthAmount        | string  | N     |        | 无注释         | format: string |
| ├─ postedAmount         | string  | N     |        | 无注释         | format: string |
| ├─ currency             | string  | N     |        | 无注释         | format: string |
| ├─ originalCurrencyCode | string  | N     |        | 原始货币代码     | format: string |
| ├─ transactionAmountInOriginalCurrency | string | N | | 原始货币中的交易金额 | format: string |
| ├─ reversalFlag         | string  | N     |        | 无注释         | format: string |
| ├─ transactionTime      | string  | N     |        | 无注释         | format: string |
| ├─ authorizationTime    | string  | N     |        | 无注释         | format: string |
| ├─ acquirerId           | string  | N     |        | 无注释         | format: string |
| ├─ merchantMcc          | string  | N     |        | 无注释         | format: string |
| ├─ merchantName         | string  | N     |        | 无注释         | format: string |
| ├─ merchantAddressAddressLine1 | string | N | | 无注释         | format: string |
| ├─ merchantAddressCity  | string  | N     |        | 无注释         | format: string |
| ├─ merchantAddressState | string  | N     |        | 无注释         | format: string |
| ├─ merchantAddressCountry | string | N | | 无注释         | format: string |
| ├─ merchantAddressZip   | string  | N     |        | 无注释         | format: string |
| ├─ posAcceptorId        | string  | N     |        | 无注释         | format: string |
| ├─ posAcceptLocation    | string  | N     |        | 无注释         | format: string |
| ├─ posEntryDescription  | string  | N     |        | 无注释         | format: string |

**响应**

| 名称     | 类型 | 必需 | 默认值 | 备注 | 其他 |
| :------- | :--- | :--- | :----- | :--- | :--- |
| (未提供)   |      |      |        |      |      |

---

**翻译结束**