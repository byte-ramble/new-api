-- OmniRouter 品牌追加 SQL 种子脚本
--
-- 目的：把 OmniRouter 品牌信息写入 new-api 的 `options` 表，无需人工进 admin 后台。
-- Rule 5 合规：保留 new-api / QuantumNous 原标识，仅追加 OmniRouter。
--
-- 适配三种数据库：PostgreSQL / MySQL / SQLite。
-- 选对应段执行；不要混跑。

-- ============================================================================
-- PostgreSQL
-- ============================================================================
-- 执行：psql -U root -d new-api -f brand-seed.sql

INSERT INTO options ("key", value) VALUES
  ('SystemName',     'OmniRouter'),
  ('Logo',           'https://cdn.omnirouter.org/brand/logo.png'),
  ('ServerAddress',  'https://omnirouter.org'),
  ('Footer',
    '<div style="text-align:center; padding:8px; font-size:12px; color:#777;">' ||
    '<strong>OmniRouter</strong> · <a href="https://omnirouter.org" target="_blank" rel="noopener">omnirouter.org</a>' ||
    ' · Powered by <a href="https://github.com/QuantumNous/new-api" target="_blank" rel="noopener">new-api</a> by QuantumNous' ||
    '</div>'
  ),
  ('HomePageContent',
    '<h1 style="text-align:center;margin-top:40px;">Welcome to OmniRouter</h1>' ||
    '<p style="text-align:center;color:#666;">One unified API for 28+ AI model groups · Pay-as-you-go in CNY · No VPN required</p>' ||
    '<p style="text-align:center;font-size:11px;color:#aaa;margin-top:60px;">Powered by new-api by QuantumNous</p>'
  ),
  ('About',
    '<h2>About OmniRouter</h2>' ||
    '<p>OmniRouter is a commercial AI API gateway aggregating 28+ model groups (Claude, GPT, Gemini, DeepSeek, Qwen, GLM, Kimi, Suno, Sora, etc.) behind a unified OpenAI / Anthropic / Gemini compatible API.</p>' ||
    '<h3>Built on</h3>' ||
    '<p>OmniRouter is built on the open-source <a href="https://github.com/QuantumNous/new-api">new-api</a> project by <strong>QuantumNous</strong>. We acknowledge and preserve all upstream attributions.</p>'
  )
ON CONFLICT ("key") DO UPDATE SET value = EXCLUDED.value;

-- ============================================================================
-- MySQL  (>=5.7.8)
-- ============================================================================
-- 执行：mysql -u root -p new-api < brand-seed.sql
--
-- 注意：MySQL 用反引号 `key` 而非双引号。
/*
INSERT INTO options (`key`, value) VALUES
  ('SystemName',     'OmniRouter'),
  ('Logo',           'https://cdn.omnirouter.org/brand/logo.png'),
  ('ServerAddress',  'https://omnirouter.org'),
  ('Footer',
    CONCAT(
      '<div style="text-align:center; padding:8px; font-size:12px; color:#777;">',
      '<strong>OmniRouter</strong> · <a href="https://omnirouter.org" target="_blank" rel="noopener">omnirouter.org</a>',
      ' · Powered by <a href="https://github.com/QuantumNous/new-api" target="_blank" rel="noopener">new-api</a> by QuantumNous',
      '</div>'
    )
  ),
  ('HomePageContent',
    CONCAT(
      '<h1 style="text-align:center;margin-top:40px;">Welcome to OmniRouter</h1>',
      '<p style="text-align:center;color:#666;">One unified API for 28+ AI model groups · Pay-as-you-go in CNY · No VPN required</p>',
      '<p style="text-align:center;font-size:11px;color:#aaa;margin-top:60px;">Powered by new-api by QuantumNous</p>'
    )
  ),
  ('About',
    CONCAT(
      '<h2>About OmniRouter</h2>',
      '<p>OmniRouter is a commercial AI API gateway aggregating 28+ model groups (Claude, GPT, Gemini, DeepSeek, Qwen, GLM, Kimi, Suno, Sora, etc.) behind a unified OpenAI / Anthropic / Gemini compatible API.</p>',
      '<h3>Built on</h3>',
      '<p>OmniRouter is built on the open-source <a href="https://github.com/QuantumNous/new-api">new-api</a> project by <strong>QuantumNous</strong>. We acknowledge and preserve all upstream attributions.</p>'
    )
  )
ON DUPLICATE KEY UPDATE value = VALUES(value);
*/

-- ============================================================================
-- SQLite
-- ============================================================================
-- 执行：sqlite3 /data/one-api.db < brand-seed.sql
/*
INSERT INTO options ("key", value) VALUES
  ('SystemName',     'OmniRouter'),
  ('Logo',           'https://cdn.omnirouter.org/brand/logo.png'),
  ('ServerAddress',  'https://omnirouter.org'),
  ('Footer',
    '<div style="text-align:center; padding:8px; font-size:12px; color:#777;">' ||
    '<strong>OmniRouter</strong> · <a href="https://omnirouter.org" target="_blank" rel="noopener">omnirouter.org</a>' ||
    ' · Powered by <a href="https://github.com/QuantumNous/new-api" target="_blank" rel="noopener">new-api</a> by QuantumNous' ||
    '</div>'
  ),
  ('HomePageContent',
    '<h1 style="text-align:center;margin-top:40px;">Welcome to OmniRouter</h1>' ||
    '<p style="text-align:center;color:#666;">One unified API for 28+ AI model groups · Pay-as-you-go in CNY · No VPN required</p>' ||
    '<p style="text-align:center;font-size:11px;color:#aaa;margin-top:60px;">Powered by new-api by QuantumNous</p>'
  ),
  ('About',
    '<h2>About OmniRouter</h2>' ||
    '<p>OmniRouter is a commercial AI API gateway aggregating 28+ model groups behind a unified OpenAI / Anthropic / Gemini compatible API.</p>' ||
    '<h3>Built on</h3>' ||
    '<p>OmniRouter is built on the open-source new-api project by QuantumNous. We acknowledge and preserve all upstream attributions.</p>'
  )
ON CONFLICT("key") DO UPDATE SET value = excluded.value;
*/

-- ============================================================================
-- 验证查询（任一数据库）
-- ============================================================================
-- SELECT "key", value FROM options WHERE "key" IN ('SystemName','Logo','Footer','HomePageContent','About','ServerAddress');
