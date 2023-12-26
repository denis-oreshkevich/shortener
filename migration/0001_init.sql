-- +goose Up
create extension if not exists "uuid-ossp";

create schema if not exists courses;

create table if not exists courses.shortener();

alter table courses.shortener add column if not exists id uuid primary key default uuid_generate_v4();
alter table courses.shortener add column if not exists short_url varchar(8) unique not null;
alter table courses.shortener add column if not exists original_url varchar unique not null;
alter table courses.shortener add column if not exists user_id uuid not null;
alter table courses.shortener add column if not exists is_deleted boolean not null default false;
-- +goose Down