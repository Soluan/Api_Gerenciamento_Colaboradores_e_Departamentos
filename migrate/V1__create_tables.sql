-- Ativa a extensão para UUIDs se ainda não estiver ativa
-- Embora a v7 seja gerada pela app, a v4 pode ser útil.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de Departamentos
-- Note que `gerente_id` e `departamento_superior_id` são nulos inicialmente
-- para permitir a inserção e evitar dependência circular na criação.
CREATE TABLE departamentos (
                               id UUID PRIMARY KEY,
                               nome VARCHAR(255) NOT NULL,
                               gerente_id UUID,
                               departamento_superior_id UUID,

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               deleted_at TIMESTAMPTZ,

                               CONSTRAINT fk_depto_superior
                                   FOREIGN KEY(departamento_superior_id)
                                       REFERENCES departamentos(id)
                                       ON DELETE SET NULL
);

-- Tabela de Colaboradores
CREATE TABLE colaboradores (
                               id UUID PRIMARY KEY,
                               nome VARCHAR(255) NOT NULL,
                               cpf VARCHAR(11) NOT NULL,
                               rg VARCHAR(20),
                               departamento_id UUID NOT NULL,

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               deleted_at TIMESTAMPTZ,

    -- Constraints de Unicidade
                               CONSTRAINT uq_cpf UNIQUE(cpf),
                               CONSTRAINT uq_rg UNIQUE(rg), -- Permite múltiplos nulos, mas único se informado

    -- Chave Estrangeira para Departamento
                               CONSTRAINT fk_colab_depto
                                   FOREIGN KEY(departamento_id)
                                       REFERENCES departamentos(id)
                                       ON DELETE RESTRICT -- Impede deletar depto com colaboradores
);

-- Adiciona a Chave Estrangeira de Gerente (circular)
-- O gerente é um colaborador. Se o gerente for demitido (deletado),
-- o departamento fica sem gerente (SET NULL).
ALTER TABLE departamentos
    ADD CONSTRAINT fk_depto_gerente
        FOREIGN KEY(gerente_id)
            REFERENCES colaboradores(id)
            ON DELETE SET NULL;

-- Índices para otimização de buscas
CREATE INDEX idx_colab_depto_id ON colaboradores(departamento_id);
CREATE INDEX idx_colab_cpf ON colaboradores(cpf);
CREATE INDEX idx_depto_superior_id ON departamentos(departamento_superior_id);
CREATE INDEX idx_depto_gerente_id ON departamentos(gerente_id);

-- Constraint para garantir que o RG, se não for nulo, é único
-- A constraint UNIQUE(rg) já faz isso, mas isso é mais explícito
-- para algumas versões de PG. A de cima é suficiente.
-- CREATE UNIQUE INDEX idx_rg_not_null ON colaboradores(rg) WHERE rg IS NOT NULL;
