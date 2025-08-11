-- Создание баз данных для Temporal
CREATE DATABASE temporal;
CREATE DATABASE temporal_visibility;

-- Подключение к базе temporal
\c temporal;

-- Создание пользователя postgres если не существует
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'postgres') THEN
        CREATE ROLE postgres WITH LOGIN PASSWORD 'password';
    END IF;
END
$$;

-- Подключение к базе temporal_visibility
\c temporal_visibility;

-- Создание пользователя postgres если не существует
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'postgres') THEN
        CREATE ROLE postgres WITH LOGIN PASSWORD 'password';
    END IF;
END
$$;
