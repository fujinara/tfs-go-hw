drop database if exists ordersdb;

create database ordersdb owner postgres;
\connect postgres

create table ordersinfo (
    orderID       varchar(100) not null,
    instrument    varchar(100) not null,
    price         float        not null,
    size          int          not null,
    order_type    varchar(100) not null,
    ts            timestamp    not null
);

alter table ordersinfo
    owner to postgres;