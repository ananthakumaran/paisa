-- Notified Cost Inflation Index values used by the debt fund transactions.
DELETE FROM ciis;
INSERT INTO ciis (financial_year, cost_inflation_index) VALUES
  ('2021-22', 317),
  ('2022-23', 331),
  ('2023-24', 348),
  ('2024-25', 363),
  ('2025-26', 376),
  ('2026-27', 384);
