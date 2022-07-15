create table dts_user_permission
(
    id serial primary key,
    code varchar not null,
    name varchar not null,
    description varchar not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp not null
);

create table dts_user_role
(
    id serial primary key,
    number integer not null,
    name varchar not null,
    description varchar not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp not null
);

create table dts_user_role_permission
(
    id integer  primary key,
    role_id integer,
    permission_id integer,
    created_at timestamp not null,
    updated_at timestamp not null
);

create table dts_user_list
(
    id serial primary key,
    role_id integer,
    department_id integer,
    number varchar not null,
    username varchar(100) not null,
    user_pass varchar(100) not null,
    email varchar(100),
    created_at timestamp not null,
    updated_at timestamp not null
);
