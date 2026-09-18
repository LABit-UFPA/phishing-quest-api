-- items generaliza phishing_emails (hoje so mockado no front) para
-- suportar multiplos canais de golpe (email, SMS, WhatsApp, site,
-- ligacao, Pix/QR), conforme ROADMAP_PESQUISA_2027.md Fase 2 e 4.
--
-- O conteudo especifico de cada canal (assunto+corpo de email, texto
-- de SMS, print de conversa, etc) fica em content_json, pois o shape
-- varia por canal e nao ha necessidade de normalizar em colunas.

CREATE TABLE IF NOT EXISTS phishing_quest.items (
    id UUID PRIMARY KEY,
    channel VARCHAR(32) NOT NULL,
    is_malicious BOOLEAN NOT NULL,
    locale VARCHAR(8) NOT NULL DEFAULT 'pt-BR',
    phish_scale_cue_count INT,
    phish_scale_premise_alignment VARCHAR(32),
    difficulty_calibrated VARCHAR(16),
    content_json JSONB NOT NULL,
    explanation TEXT,
    source VARCHAR(64),
    reviewed_by UUID REFERENCES phishing_quest.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_items_channel CHECK (channel IN ('email', 'sms', 'whatsapp', 'website', 'phone_call', 'pix_qr'))
);

COMMENT ON TABLE phishing_quest.items IS 'Item de simulacao (email, SMS, WhatsApp, site falso, ligacao, Pix/QR) usado no jogo e no estudo de retencao. Generaliza o antigo conceito de phishing_email (que so cobria email) para multicanal.';
COMMENT ON COLUMN phishing_quest.items.channel IS 'Canal simulado: email | sms | whatsapp | website | phone_call | pix_qr';
COMMENT ON COLUMN phishing_quest.items.is_malicious IS 'true = item e uma tentativa de golpe; false = item legitimo (necessario para medir falso alarme)';
COMMENT ON COLUMN phishing_quest.items.phish_scale_cue_count IS 'Contagem de pistas segundo o NIST Phish Scale, usada para calibrar dificuldade';
COMMENT ON COLUMN phishing_quest.items.phish_scale_premise_alignment IS 'Alinhamento da premissa do NIST Phish Scale com o contexto do usuario (ex.: low, medium, high)';
COMMENT ON COLUMN phishing_quest.items.difficulty_calibrated IS 'Dificuldade calibrada a partir do Phish Scale (ex.: easy, medium, hard), substitui o enum arbitrario usado no front hoje';
COMMENT ON COLUMN phishing_quest.items.content_json IS 'Conteudo especifico do canal (assunto/corpo de email, texto de SMS, mensagens de WhatsApp, etc). Shape varia por canal.';
COMMENT ON COLUMN phishing_quest.items.reviewed_by IS 'Usuario (admin/researcher) que revisou o item antes de publicacao — obrigatorio para itens gerados por LLM (issue #30)';

CREATE INDEX IF NOT EXISTS idx_items_channel ON phishing_quest.items(channel);
CREATE INDEX IF NOT EXISTS idx_items_is_malicious ON phishing_quest.items(is_malicious);
