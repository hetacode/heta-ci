 CREATE DATABASE "heta-ci"
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;

\c "heta-ci";

-- Table: public.kv_build_last_commit

-- DROP TABLE public.kv_build_last_commit;

CREATE TABLE IF NOT EXISTS public.kv_build_last_commit
(
    id SERIAL PRIMARY KEY,
    key character varying(150) COLLATE pg_catalog."default" NOT NULL,
    value_hash_commit character varying(40) COLLATE pg_catalog."default" NOT NULL,
    created_at bigint NOT NULL
)

TABLESPACE pg_default;

ALTER TABLE public.kv_build_last_commit
    OWNER to postgres;
    
-- Table: public.build

-- DROP TABLE public.build;

CREATE TABLE IF NOT EXISTS public.build
(
    uid uuid NOT NULL,
    repository_hash character varying(65) COLLATE pg_catalog."default" NOT NULL,
    commit_hash character varying(40) COLLATE pg_catalog."default" NOT NULL,
    pipeline_json json NOT NULL,
    logs text COLLATE pg_catalog."default",
    result_status character varying(10) COLLATE pg_catalog."default" NOT NULL,
    artifacts_uid uuid,
    created_at bigint NOT NULL,
    finish_at bigint,
    CONSTRAINT build_pkey PRIMARY KEY (uid)
)

TABLESPACE pg_default;

ALTER TABLE public.build
    OWNER to postgres;
	
-- Table: public.project

-- DROP TABLE public.project;

CREATE TABLE IF NOT EXISTS public.project
(
    uid uuid NOT NULL,
    name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    description text COLLATE pg_catalog."default",
    created_at bigint NOT NULL,
    CONSTRAINT project_pkey PRIMARY KEY (uid)
)

TABLESPACE pg_default;

ALTER TABLE public.project
    OWNER to postgres;
	
-- Table: public.repository

-- DROP TABLE public.repository;

CREATE TABLE IF NOT EXISTS public.repository
(
    repo_hash character varying(65) COLLATE pg_catalog."default" NOT NULL,
    repository_url text COLLATE pg_catalog."default" NOT NULL,
    default_branch text COLLATE pg_catalog."default" NOT NULL,
    name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    created_at bigint NOT NULL,
    project_uid uuid NOT NULL,
    CONSTRAINT repository_pkey PRIMARY KEY (repo_hash)
)

TABLESPACE pg_default;

ALTER TABLE public.repository
    OWNER to postgres;