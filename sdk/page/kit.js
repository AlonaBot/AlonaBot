/**
 * kit.js: AlonaBot 插件页面套件客户端 (ADMIN-API §7 / DECISIONS D-84, D-159)
 *
 * 在 window.AlonaBot 上提供与 WebUI postMessage 桥通信的接口：
 * - request(method, path, body?) → Promise<Response 数据>
 * - t(key, args?) → Promise<string>
 * - onTheme(fn)
 * - onLocale(fn)
 */
(function(window) {
  'use strict';

  var CHANNEL = 'alonabot.plugin';

  var DEFAULT_TOKENS = {
    '--ab-bg': '#F3F6F9',
    '--ab-surface': '#FFFFFF',
    '--ab-surface-2': '#F4F7FA',
    '--ab-line': '#DFE5EC',
    '--ab-text': '#1F2A44',
    '--ab-text-muted': '#5E6B82',
    '--ab-accent': '#0E7C92',
    '--ab-accent-soft': '#E2F5F8',
    '--ab-success': '#1F9D6B',
    '--ab-warning': '#D99A06',
    '--ab-danger': '#F4505A',
    '--ab-critical': '#C084FC',
    '--ab-radius-sm': '6px',
    '--ab-radius': '10px',
    '--ab-radius-lg': '16px',
    '--ab-font-sans': '"PingFang SC", "Microsoft YaHei", "Noto Sans CJK SC", "Noto Sans SC", system-ui, sans-serif',
    '--ab-font-mono': '"JetBrains Mono", ui-monospace, monospace'
  };

  var inIframe = (function() {
    try {
      return window.self !== window.top && window.parent && window.parent !== window;
    } catch (e) {
      return true;
    }
  })();

  var currentTheme = 'light';
  var currentLocale = 'en-US';
  var currentTokens = Object.assign({}, DEFAULT_TOKENS);
  var themeListeners = [];
  var localeListeners = [];
  var pendingRequests = new Map();

  var messageCounter = 0;

  function nextId() {
    messageCounter = (messageCounter + 1) & 0x7fffffff;
    return 'ab-' + Date.now().toString(36) + '-' + messageCounter.toString(36);
  }

  function applyTokens(tokens) {
    if (!tokens || typeof tokens !== 'object') return;
    var root = document.documentElement;
    if (!root) return;
    for (var key in tokens) {
      if (Object.prototype.hasOwnProperty.call(tokens, key)) {
        if (typeof key === 'string' && key.indexOf('--ab-') === 0 && typeof tokens[key] === 'string') {
          currentTokens[key] = tokens[key];
          root.style.setProperty(key, tokens[key]);
        }
      }
    }
  }

  function applyTheme(theme) {
    if (!theme || typeof theme !== 'string') return;
    currentTheme = theme;
    var root = document.documentElement;
    if (root) {
      root.setAttribute('data-theme', theme);
    }
  }

  function notifyThemeListeners() {
    var snapshot = Object.assign({}, currentTokens);
    for (var i = 0; i < themeListeners.length; i++) {
      try {
        themeListeners[i](currentTheme, snapshot);
      } catch (err) {
        console.error('AlonaBot onTheme error:', err);
      }
    }
  }

  function applyThemeAndTokens(theme, tokens) {
    if (theme) applyTheme(theme);
    if (tokens) applyTokens(tokens);
    notifyThemeListeners();
  }

  function applyLocale(locale) {
    currentLocale = locale;
    notifyLocaleListeners();
  }

  function notifyLocaleListeners() {
    for (var i = 0; i < localeListeners.length; i++) {
      try {
        localeListeners[i](currentLocale);
      } catch (err) {
        console.error('AlonaBot onLocale error:', err);
      }
    }
  }

  function postToParent(message) {
    if (!inIframe) return;
    window.parent.postMessage(message, '*');
  }

  function sendBridgeRequest(op, payload) {
    if (!inIframe) {
      return Promise.reject(new Error('Not running inside iframe; bridge unavailable'));
    }
    var id = nextId();
    var message = Object.assign({
      channel: CHANNEL,
      id: id,
      op: op
    }, payload || {});

    return new Promise(function(resolve, reject) {
      pendingRequests.set(id, { op: op, resolve: resolve, reject: reject });
      try {
        postToParent(message);
      } catch (err) {
        pendingRequests.delete(id);
        reject(err);
      }
    });
  }

  function utf8ToBase64(str) {
    var bytes = new TextEncoder().encode(str);
    var binary = '';
    var len = bytes.byteLength;
    for (var i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  function base64ToUtf8(b64) {
    if (!b64) return '';
    var binary = atob(b64);
    var len = binary.length;
    var bytes = new Uint8Array(len);
    for (var i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return new TextDecoder().decode(bytes);
  }

  function request(method, path, body) {
    if (!inIframe) {
      return Promise.reject(new Error('AlonaBot bridge request is only available inside iframe'));
    }

    var httpMethod = (method || 'GET').toUpperCase();
    var payload = {
      method: httpMethod,
      path: path,
      headers: {}
    };

    if (body !== undefined && body !== null) {
      var bodyStr = '';
      if (typeof body === 'object') {
        bodyStr = JSON.stringify(body);
        payload.headers['Content-Type'] = 'application/json';
      } else {
        bodyStr = String(body);
      }
      payload.body_base64 = utf8ToBase64(bodyStr);
    }

    return sendBridgeRequest('request', payload).then(function(result) {
      var bodyText = result.body_base64 ? base64ToUtf8(result.body_base64) : '';
      var data = bodyText;
      var headers = result.headers || {};
      var contentType = (headers['content-type'] || headers['Content-Type'] || '').toLowerCase();
      if (contentType.indexOf('application/json') >= 0 || contentType.indexOf('+json') >= 0) {
        try {
          data = JSON.parse(bodyText);
        } catch (e) {}
      } else if (bodyText.indexOf('{') === 0 || bodyText.indexOf('[') === 0) {
        try {
          data = JSON.parse(bodyText);
        } catch (e) {}
      }

      return {
        status: result.status,
        ok: result.status >= 200 && result.status < 300,
        headers: headers,
        body: bodyText,
        data: data,
        body_base64: result.body_base64,
        text: function() { return Promise.resolve(bodyText); },
        json: function() { return Promise.resolve(data); }
      };
    });
  }

  function t(key, args) {
    if (!inIframe) return Promise.resolve(key);
    return sendBridgeRequest('translate', { key: key, args: args }).then(function(result) {
      return result.text;
    });
  }

  function onTheme(fn) {
    if (typeof fn !== 'function') return function() {};
    themeListeners.push(fn);
    try {
      fn(currentTheme, Object.assign({}, currentTokens));
    } catch (err) {
      console.error('AlonaBot onTheme error:', err);
    }
    return function unsubscribe() {
      var idx = themeListeners.indexOf(fn);
      if (idx >= 0) themeListeners.splice(idx, 1);
    };
  }

  function onLocale(fn) {
    if (typeof fn !== 'function') return function() {};
    localeListeners.push(fn);
    try {
      fn(currentLocale);
    } catch (err) {
      console.error('AlonaBot onLocale error:', err);
    }
    return function unsubscribe() {
      var idx = localeListeners.indexOf(fn);
      if (idx >= 0) localeListeners.splice(idx, 1);
    };
  }

  // 快捷键捕获 (⌘K / Ctrl+K 转发为 shortcut op)
  window.addEventListener('keydown', function(event) {
    if ((event.metaKey || event.ctrlKey) && (event.key === 'k' || event.key === 'K')) {
      event.preventDefault();
      if (inIframe) {
        postToParent({
          channel: CHANNEL,
          id: nextId(),
          op: 'shortcut',
          key: 'mod+k'
        });
      }
    }
  }, true);

  function isRecord(value) {
    return value !== null && typeof value === 'object' && !Array.isArray(value);
  }

  function isStringRecord(value) {
    return isRecord(value) && Object.keys(value).every(function(key) {
      return typeof value[key] === 'string';
    });
  }

  function isThemeData(value) {
    return isRecord(value) && (value.theme === 'light' || value.theme === 'dark') && isStringRecord(value.tokens);
  }

  function isResult(op, result) {
    if (!isRecord(result)) return false;
    switch (op) {
      case 'hello':
        return typeof result.locale === 'string' && isThemeData(result);
      case 'translate':
        return typeof result.text === 'string';
      case 'request':
        return typeof result.status === 'number' && isStringRecord(result.headers) && typeof result.body_base64 === 'string';
    }
    return false;
  }

  // 沙箱页面使用不透明源，按窗口身份匹配父窗口。
  window.addEventListener('message', function(event) {
    if (!inIframe || event.source !== window.parent) {
      return;
    }

    var data = event.data;
    if (!isRecord(data) || data.channel !== CHANNEL) {
      return;
    }

    // 响应消息处理
    if (Object.prototype.hasOwnProperty.call(data, 'id')) {
      if (typeof data.id !== 'string' || typeof data.ok !== 'boolean') return;
      var pending = pendingRequests.get(data.id);
      if (pending) {
        if (data.ok) {
          if (!isResult(pending.op, data.result)) return;
        } else if (!isRecord(data.error) || typeof data.error.code !== 'string' || typeof data.error.safe_message !== 'string') {
          return;
        }
        pendingRequests.delete(data.id);
        if (data.ok) {
          pending.resolve(data.result);
        } else {
          var err = new Error(data.error.safe_message);
          err.code = data.error.code;
          pending.reject(err);
        }
      }
      return;
    }

    // 服务端推送事件处理
    if (isRecord(data.data)) {
      if (data.event === 'theme' && isThemeData(data.data)) {
        applyThemeAndTokens(data.data.theme, data.data.tokens);
      } else if (data.event === 'locale' && typeof data.data.locale === 'string') {
        applyLocale(data.data.locale);
      }
    }
  });

  // 初始化：应用默认浅色样式，并在 iframe 环境中发送 hello 握手
  applyThemeAndTokens('light', DEFAULT_TOKENS);
  if (inIframe) {
    sendBridgeRequest('hello')
      .then(function(result) {
        applyThemeAndTokens(result.theme, result.tokens);
        applyLocale(result.locale);
      })
      .catch(function(err) {
        console.error('AlonaBot hello failed:', err);
      });
  }

  var AlonaBot = {
    request: request,
    t: t,
    onTheme: onTheme,
    onLocale: onLocale
  };

  window.AlonaBot = AlonaBot;
})(typeof window !== 'undefined' ? window : this);
