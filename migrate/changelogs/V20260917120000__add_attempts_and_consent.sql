-- attempts e o dado central do artigo (ROADMAP_PESQUISA_2027.md, Fase 2):
-- veredito, acao, confianca, justificativa, latencia, clique por
-- tentativa. POST /attempts exige consentimento previo do usuario, daqui
-- a necessidade de uma tabela minima de consentimento nesta migration.
--
-- study_participants aqui e uma versao MINIMA (so o suficiente para
-- validar consentimento antes de registrar uma tentativa). A issue #25
-- vai expandir esta tabela via ALTER TABLE com os campos completos do
-- desenho experimental: cohort_id, condition, consent_version,
-- demographics_json, withdrawn_at. Nao antecipamos esses campos aqui
-- para nao definir o desenho experimental (condicoes, versionamento de
-- TCLE) fora do escopo desta issue.

CREATE TABLE IF NOT EXISTS phishing_quest.study_participants (
    user_id UUID PRIMARY KEY REFERENCES phishing_quest.users(id),
    consented_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE phishing_quest.study_participants IS 'Registro minimo de consentimento do usuario para participar da coleta de dados de pesquisa. Expandida pela issue #25 com cohort/condition/demographics/withdrawn_at.';

CREATE TABLE IF NOT EXISTS phishing_quest.attempts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES phishing_quest.users(id),
    item_id UUID NOT NULL REFERENCES phishing_quest.items(id),
    session_id UUID NOT NULL,
    condition VARCHAR(64),
    verdict BOOLEAN,
    action VARCHAR(32) NOT NULL,
    confidence INT,
    justification TEXT,
    llm_rating INT,
    llm_feedback_json JSONB,
    is_correct BOOLEAN,
    latency_ms INT,
    clicked_link BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_attempts_action CHECK (action IN ('report', 'delete', 'verify_other_channel', 'reply', 'click', 'ignore')),
    CONSTRAINT chk_attempts_confidence CHECK (confidence IS NULL OR (confidence BETWEEN 1 AND 5))
);

COMMENT ON TABLE phishing_quest.attempts IS 'Registro completo de uma tentativa: veredito do usuario, acao escolhida, confianca, justificativa, avaliacao por IA, latencia de decisao e se houve clique em link malicioso. E o dado central do estudo de retencao.';
COMMENT ON COLUMN phishing_quest.attempts.condition IS 'Condicao experimental atribuida ao usuario (ex.: feedback_formativo, feedback_binario) — referencia study_participants.condition apos a issue #25';
COMMENT ON COLUMN phishing_quest.attempts.verdict IS 'true = usuario julgou o item como malicioso; false = julgou legitimo; NULL se a interacao nao envolveu veredito binario';
COMMENT ON COLUMN phishing_quest.attempts.action IS 'Acao escolhida: report | delete | verify_other_channel | reply | click | ignore';
COMMENT ON COLUMN phishing_quest.attempts.confidence IS 'Escala de confianca de 1 a 5 informada pelo usuario';
COMMENT ON COLUMN phishing_quest.attempts.llm_feedback_json IS 'Feedback formativo gerado por LLM a partir da justificativa do usuario (strengths, improvements, etc)';
COMMENT ON COLUMN phishing_quest.attempts.is_correct IS 'Se o veredito/acao do usuario estava correto para este item (usa items.is_malicious)';
COMMENT ON COLUMN phishing_quest.attempts.latency_ms IS 'Tempo em milissegundos entre a exibicao do item e a decisao do usuario';
COMMENT ON COLUMN phishing_quest.attempts.clicked_link IS 'Se o usuario clicou em algum link presente no item durante a tentativa';

CREATE INDEX IF NOT EXISTS idx_attempts_user_id ON phishing_quest.attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_attempts_item_id ON phishing_quest.attempts(item_id);
CREATE INDEX IF NOT EXISTS idx_attempts_session_id ON phishing_quest.attempts(session_id);
