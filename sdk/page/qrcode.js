(function(window) {
  'use strict';

  // QR Code Model 2 的 M 级纠错参数，按版本 1 至 40 排列。
  const parityLengths = [10, 16, 26, 18, 24, 16, 18, 22, 22, 26, 30, 22, 22, 24, 24, 28, 28, 26, 26, 26,
    26, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28];
  const blockCounts = [1, 1, 1, 2, 2, 4, 4, 4, 5, 5, 5, 8, 9, 9, 10, 10, 11, 13, 14, 16,
    17, 17, 18, 20, 21, 23, 25, 26, 28, 29, 31, 33, 35, 37, 38, 40, 43, 45, 47, 49];
  const masks = [
    (x, y) => (x + y) % 2 === 0,
    (x, y) => y % 2 === 0,
    (x, y) => x % 3 === 0,
    (x, y) => (x + y) % 3 === 0,
    (x, y) => (Math.floor(y / 2) + Math.floor(x / 3)) % 2 === 0,
    (x, y) => x * y % 2 + x * y % 3 === 0,
    (x, y) => (x * y % 2 + x * y % 3) % 2 === 0,
    (x, y) => ((x + y) % 2 + x * y % 3) % 2 === 0
  ];

  function totalCodewords(version) {
    let modules = (16 * version + 128) * version + 64;
    if (version >= 2) {
      const count = Math.floor(version / 7) + 2;
      modules -= (25 * count - 10) * count - 55;
      if (version >= 7) modules -= 36;
    }
    return Math.floor(modules / 8);
  }

  function multiply(a, b) {
    let product = 0;
    for (; b; b >>>= 1) {
      if (b & 1) product ^= a;
      a = (a << 1) ^ (a & 0x80 ? 0x11d : 0);
    }
    return product;
  }

  function interleave(data, count, parityLength) {
    // 生成多项式的根为 GF(256) 中连续的 2 的幂，从 2^0 开始。
    let divisor = [1];
    for (let root = 1; divisor.length <= parityLength; root = multiply(root, 2)) {
      divisor = [1, ...divisor.map((coefficient, i) => multiply(coefficient, root) ^ (divisor[i + 1] || 0))];
    }
    const chunks = [], parity = [];
    const shortLength = Math.floor(data.length / count);
    const shortCount = count - data.length % count;
    for (let i = 0, offset = 0; i < count; i++) {
      const length = shortLength + Number(i >= shortCount);
      const chunk = data.slice(offset, offset + length);
      offset += length;
      const remainder = chunk.concat(Array(parityLength).fill(0));
      for (let j = 0; j < length; j++) {
        const lead = remainder[j];
        for (let k = 1; k <= parityLength; k++) remainder[j + k] ^= multiply(lead, divisor[k]);
      }
      chunks.push(chunk);
      parity.push(remainder.slice(length));
    }

    // 短块先于长块；按列交错数据，再按列交错纠错码字。
    const result = [];
    for (const blocks of [chunks, parity]) {
      for (let i = 0; i < blocks[count - 1].length; i++) {
        for (const block of blocks) if (i < block.length) result.push(block[i]);
      }
    }
    return result;
  }

  function encode(text) {
    // 1. 按 UTF-8 字节选最小版本；非 ASCII 文本用 ECI 26 声明编码。
    const bytes = new TextEncoder().encode(text);
    const utf8 = bytes.some(byte => byte >= 0x80);
    for (let version = 1; version <= 40; version++) {
      const parityLength = parityLengths[version - 1], count = blockCounts[version - 1];
      const capacity = totalCodewords(version) - parityLength * count;
      const lengthBits = version < 10 ? 8 : 16;
      if ((utf8 ? 12 : 0) + 4 + lengthBits + 8 * bytes.length > 8 * capacity) continue;

      // 2. 写入模式、字节长度与数据，补终止位及填充码字后生成纠错块。
      const bits = [];
      function append(value, width) {
        for (let i = width - 1; i >= 0; i--) bits.push((value >>> i) & 1);
      }
      if (utf8) append(0x71a, 12); // ECI 模式 0111，UTF-8 编号 26。
      append(0b0100, 4);
      append(bytes.length, lengthBits);
      for (const byte of bytes) append(byte, 8);
      append(0, Math.min(4, 8 * capacity - bits.length));
      append(0, (8 - bits.length % 8) % 8);
      const data = [];
      for (let i = 0; i < bits.length; i += 8) {
        data.push(bits.slice(i, i + 8).reduce((value, bit) => (value << 1) | bit, 0));
      }
      for (let pad = 0; data.length < capacity; pad++) data.push(pad % 2 ? 0x11 : 0xec);
      return drawSymbol(version, interleave(data, count, parityLength));
    }
    throw new RangeError('Text exceeds QR Code byte capacity at error correction level M');
  }

  function bch(value, degree, polynomial) {
    let remainder = value << degree;
    for (let i = 31 - Math.clz32(remainder); i >= degree; i--) {
      if ((remainder >>> i) & 1) remainder ^= polynomial << (i - degree);
    }
    return (value << degree) | remainder;
  }

  function drawSymbol(version, codewords) {
    const size = version * 4 + 17;
    // -1 表示待填数据格；定位、时序、对齐与格式格不参与遮罩。
    const symbol = Array.from({length: size}, () => new Int8Array(size).fill(-1));
    const set = (x, y, dark) => { symbol[y][x] = Number(dark); };
    const bit = (value, i) => (value >>> i) & 1;

    // 1. 绘制功能图形；对齐图形可以覆盖时序线，但避开三个定位角。
    for (let i = 0; i < size; i++) {
      set(i, 6, i % 2 === 0);
      set(6, i, i % 2 === 0);
    }
    for (const [cx, cy] of [[3, 3], [size - 4, 3], [3, size - 4]]) {
      for (let dy = -4; dy <= 4; dy++) {
        for (let dx = -4; dx <= 4; dx++) {
          const x = cx + dx, y = cy + dy;
          const d = Math.max(Math.abs(dx), Math.abs(dy));
          if (x >= 0 && y >= 0 && x < size && y < size) set(x, y, d !== 2 && d !== 4);
        }
      }
    }
    if (version >= 2) {
      const count = Math.floor(version / 7) + 2;
      const step = version === 32 ? 26 : 2 * Math.ceil((size - 13) / (2 * (count - 1)));
      const last = size - 7;
      const centers = Array.from({length: count}, (_, i) => i === 0 ? 6 : last - (count - 1 - i) * step);
      for (const cy of centers) for (const cx of centers) {
        if ((cx === 6 && (cy === 6 || cy === last)) || (cx === last && cy === 6)) continue;
        for (let dy = -2; dy <= 2; dy++) for (let dx = -2; dx <= 2; dx++) {
          set(cx + dx, cy + dy, Math.max(Math.abs(dx), Math.abs(dy)) !== 1);
        }
      }
    }
    if (version >= 7) {
      const versionBits = bch(version, 12, 0x1f25);
      for (let i = 0; i < 18; i++) {
        const x = size - 11 + i % 3, y = Math.floor(i / 3);
        set(x, y, bit(versionBits, i));
        set(y, x, bit(versionBits, i));
      }
    }

    function format(mask) {
      // M 级的格式值高两位为 00，低三位为遮罩编号。
      const value = bch(mask, 10, 0x537) ^ 0x5412;
      for (let i = 0; i < 15; i++) {
        const dark = bit(value, i);
        if (i < 8) {
          set(8, i < 6 ? i : i + 1, dark);
          set(size - 1 - i, 8, dark);
        } else {
          set(i === 8 ? 7 : 14 - i, 8, dark);
          set(8, size - 15 + i, dark);
        }
      }
      set(8, size - 8, true);
    }

    // 2. 从右下角沿双列交替向上、向下写入数据，跳过列 6 与功能格。
    format(0);
    const dataCells = [];
    for (let right = size - 1, up = true; right >= 1; right -= 2, up = !up) {
      if (right === 6) right--;
      for (let i = 0; i < size; i++) {
        const y = up ? size - 1 - i : i;
        for (let x = right; x >= right - 1; x--) if (symbol[y][x] === -1) dataCells.push([x, y]);
      }
    }

    // 3. 比较八种遮罩，选扫描干扰惩罚最低的一种。
    let best, lowest = Infinity;
    for (let mask = 0; mask < masks.length; mask++) {
      format(mask);
      dataCells.forEach(([x, y], i) => {
        const dark = i < codewords.length * 8 ? bit(codewords[i >>> 3], 7 - i % 8) : 0;
        set(x, y, dark ^ Number(masks[mask](x, y)));
      });
      const score = penalty(symbol);
      if (score < lowest) {
        best = symbol.map(row => row.slice());
        lowest = score;
      }
    }
    return best;
  }

  function penalty(symbol) {
    const size = symbol.length;
    let score = 0, dark = 0;
    for (let i = 0; i < size; i++) {
      for (const line of [symbol[i], symbol.map(row => row[i])]) {
        let run = 0, last = -1, pattern = 0;
        for (let j = 0; j < size + 4; j++) {
          const cell = j < size ? line[j] : 0;
          if (j < size) {
            run = cell === last ? run + 1 : 1;
            if (run >= 5) score += run === 5 ? 3 : 1;
            last = cell;
          }
          pattern = ((pattern << 1) | cell) & 0x7ff;
          if (j >= 6 && (pattern === 0b00001011101 || pattern === 0b10111010000)) score += 40;
        }
      }
      for (let x = 0; x < size; x++) {
        const cell = symbol[i][x];
        dark += cell;
        if (i > 0 && x > 0 && cell === symbol[i][x - 1] && cell === symbol[i - 1][x - 1] && cell === symbol[i - 1][x]) score += 3;
      }
    }
    return score + 10 * Math.floor(Math.abs(dark * 20 - size * size * 10) / (size * size));
  }

  function renderQRCode(element, text, {size = 256} = {}) {
    const symbol = encode(text);
    const quiet = 4;
    const viewSize = symbol.length + quiet * 2;
    const ns = 'http://www.w3.org/2000/svg';
    const document = element.ownerDocument;
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('viewBox', `0 0 ${viewSize} ${viewSize}`);
    svg.setAttribute('width', size);
    svg.setAttribute('height', size);
    svg.setAttribute('shape-rendering', 'crispEdges');
    const paper = document.createElementNS(ns, 'rect');
    paper.setAttribute('width', viewSize);
    paper.setAttribute('height', viewSize);
    paper.setAttribute('fill', 'var(--ab-qr-paper, #fff)');
    const ink = document.createElementNS(ns, 'path');
    const squares = [];
    symbol.forEach((row, y) => row.forEach((dark, x) => {
      if (dark) squares.push(`M${x + quiet} ${y + quiet}h1v1h-1z`);
    }));
    ink.setAttribute('d', squares.join(''));
    ink.setAttribute('fill', 'var(--ab-qr-ink, #000)');
    svg.append(paper, ink);
    element.replaceChildren(svg);
  }

  window.AlonaBot.renderQRCode = renderQRCode;
})(window);
