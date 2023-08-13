CREATE TABLE IF NOT EXISTS `shoma`.`migration`
(
    `id`        UInt64,
    `migration` String,
    `batch`     UInt32
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, migration, batch) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`users`
(
    `id`            UInt64,
    `national_code` String,
    `username` Nullable(String),
    `birthday` Nullable(Date),
    `province_id` Nullable(UInt64),
    `city_id` Nullable(UInt64),
    `company_name` Nullable(String),
    `company_id`    UInt64,
    `identify`      UInt8,
    `credit_amount` UInt64,
    `debt_amount`   UInt64,
    `is_active`     UInt8,
    `fcm_token` Nullable(String),
    `device_id` Nullable(String),
    `app_version` Nullable(String),
    `uuid`          String,
    `api_token` Nullable(String),
    `subset_id`     UInt64,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
) ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, national_code, username, province_id, city_id, company_name,
                                                company_id,
                                                identify, birthday, credit_amount, debt_amount, created_at,
                                                subset_id) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`branches`
(
    `id`                  UInt64,
    `supplier_id` Nullable(UInt64),
    `store_id`            UInt64,
    `kian_branch_id` Nullable(String),
    `name`                String,
    `en_name` Nullable(String),
    `default_installment` UInt32,
    `percentage_interest` Int8,
    `interest`            Float64,
    `city_name` Nullable(String),
    `address` Nullable(String),
    `gps_info` Nullable(String),
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, store_id, name, city_name, address) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`city`
(
    `id`          UInt64,
    `province_id` UInt64,
    `name`        String
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, province_id, name) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`installments`
(
    `id`             UInt64,
    `company_id`     UInt64,
    `supplier_id`    UInt64,
    `transaction_id` UInt64,
    `user_id`        UInt64,
    `store_id`       UInt64,
    `branch_id`      UInt64,
    `amount`         UInt64,
    `pay_status`     UInt8,
    `is_canceled`    UInt8,
    `pay_date`       Date,
    `account_number` Nullable(String),
    `settle_status`  UInt8,
    `contract_number` Nullable(String),
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
) ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, company_id, supplier_id, transaction_id, user_id, store_id,
                                                branch_id, amount,
                                                pay_status, is_canceled, pay_date, settle_status,
                                                created_at) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`mobiles`
(
    `id`          UInt64,
    `user_id`     UInt64,
    `mobile`      UInt64,
    `is_active`   UInt8,
    `is_verified` UInt8,
    `is_primary`  UInt8,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, user_id, mobile, is_active, is_verified, is_primary, created_at) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`province`
(
    `id`   UInt64,
    `name` String
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, name) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`stores`
(
    `id`                  UInt64,
    `supplier_id`         UInt64,
    `kian_store_id` Nullable(String),
    `name`                String,
    `default_installment` UInt32,
    `percentage_interest` Nullable(Int8) default 0,
    `interest` Nullable(Float64)         default 0,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, supplier_id, name, kian_store_id, default_installment) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`suppliers`
(
    `id`        UInt64,
    `name`      String,
    `phone` Nullable(String),
    `token` Nullable(String),
    `is_active` UInt8,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, name, phone, token, is_active, created_at) SETTINGS allow_nullable_key = 1,index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`transactions`
(
    `id`              UInt64,
    `user_id`         UInt64,
    `company_id`      UInt64,
    `supplier_id`     UInt64,
    `store_id`        UInt64,
    `branch_id`       UInt64,
    `amount`          UInt64,
    `details` Nullable(String),
    `credit_amount`   UInt64,
    `installment`     UInt32,
    `purchase_password` Nullable(UInt64),
    `discount`        UInt64,
    `tax`             UInt64,
    `supplier_transaction_id` Nullable(String),
    `checkout_status` UInt8,
    `is_canceled`     UInt8,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
) ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, supplier_id, company_id, user_id, store_id, branch_id, amount,
                                                details,
                                                credit_amount, installment,
                                                purchase_password, discount, tax, supplier_transaction_id,
                                                checkout_status,
                                                is_canceled,
                                                created_at) SETTINGS allow_nullable_key = 1, index_granularity = 8192, index_granularity_bytes = 0;
CREATE TABLE IF NOT EXISTS `shoma`.`transaction_detail`
(
    `id`             UInt64,
    `transaction_id` UInt64,
    `detail`         String,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, transaction_id, created_at) SETTINGS allow_nullable_key = 1;
CREATE TABLE IF NOT EXISTS `shoma`.`company`
(
    `id`   UInt64,
    `name` String,
    `created_at` Nullable(DateTime),
    `updated_at` Nullable(DateTime)
)
    ENGINE = MergeTree PRIMARY KEY (id) ORDER BY (id, name, created_at) SETTINGS allow_nullable_key = 1, index_granularity = 8192, index_granularity_bytes = 0;