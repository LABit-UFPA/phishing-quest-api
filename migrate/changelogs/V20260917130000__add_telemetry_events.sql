-- telemetry_events suporta a fila offline do app: o cliente gera o id
-- de cada evento localmente (uuid) antes de enviar, entao reenviar o
-- mesmo lote apos uma falha de rede parcial nao duplica eventos — o
-- endpoint de ingestao faz upsert (ON CONFLICT DO NOTHING) por id.

CREATE TABLE IF NOT EXISTS phishing_quest.telemetry_events (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES phishing_quest.users(id),
    session_id UUID,
    event_type VARCHAR(64) NOT NULL,
    payload_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE phishing_quest.telemetry_events IS 'Eventos de telemetria de uso do app (ex.: tela aberta, notificacao de revisao espacada respondida, erro de rede). Complementa attempts, que registra so a decisao sobre um item.';
COMMENT ON COLUMN phishing_quest.telemetry_events.id IS 'Gerado pelo cliente (nao pelo servidor) para permitir reenvio idempotente da fila offline sem duplicar eventos';
COMMENT ON COLUMN phishing_quest.telemetry_events.user_id IS 'Pode ser NULL para eventos anonimos anteriores ao login';

CREATE INDEX IF NOT EXISTS idx_telemetry_events_user_id ON phishing_quest.telemetry_events(user_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_events_session_id ON phishing_quest.telemetry_events(session_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_events_event_type ON phishing_quest.telemetry_events(event_type);
