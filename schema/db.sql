create table trs_product_category_info
(
    id serial not null
        constraint trs_product_category_info_pk
            primary key,
    category_name varchar,
    active boolean,
    created_at timestamp,
    updated_at timestamp
);