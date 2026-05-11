'use client';
import React from 'react';

type Token =
  | { type: 'text'; value: string }
  | { type: 'tagOpen'; name: string; attrs: Record<string, string> }
  | { type: 'tagClose'; name: string }
  | { type: 'selfClose'; name: string; attrs: Record<string, string> };

const SUPPORTED = new Set(['b', 'i', 'br', 'ul', 'li', 'a']);

function parseAttrs(raw: string): Record<string, string> {
  const attrs: Record<string, string> = {};
  // very small attr parser: key="value" | key='value' | key=value
  const re = /(\w+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'>]+))/g;

  let m: RegExpExecArray | null;

  while ((m = re.exec(raw))) {
    attrs[m[1]!] = m[2] ?? m[3] ?? m[4] ?? '';
  }

  return attrs;
}

function tokenize(html: string): Token[] {
  const tokens: Token[] = [];
  const re = /<\/?([a-zA-Z0-9]+)([^>]*)>|([^<]+)/g;

  let m: RegExpExecArray | null;

  while ((m = re.exec(html))) {
    // Text node
    if (m[3]) {
      tokens.push({ type: 'text', value: m[3] });
      continue;
    }

    const name = (m[1] || '').toLowerCase();
    const rawAttrs = m[2] || '';

    if (!SUPPORTED.has(name)) {
      // treat unsupported tags as text (keep original)
      tokens.push({ type: 'text', value: m[0] });
      continue;
    }

    const isClose = m[0].startsWith('</');
    const isSelf = name === 'br' || /\/\s*>$/.test(m[0]);
    const attrs = parseAttrs(rawAttrs);

    if (isSelf) {
      tokens.push({ type: 'selfClose', name, attrs });
    } else if (isClose) {
      tokens.push({ type: 'tagClose', name });
    } else {
      tokens.push({ type: 'tagOpen', name, attrs });
    }
  }

  return tokens;
}

function buildReactTree(tokens: Token[]): React.ReactNode[] {
  type Frame = {
    name: string;
    attrs?: Record<string, string>;
    children: React.ReactNode[];
    key: number;
  };
  const out: React.ReactNode[] = [];
  const stack: Frame[] = [];

  let keyCounter = 0;

  const nextKey = () => keyCounter++;

  const pushNode = (node: React.ReactNode) => {
    const top = stack[stack.length - 1];

    if (top) top.children.push(node);
    else out.push(node);
  };

  for (const t of tokens) {
    if (t.type === 'text') {
      pushNode(t.value);
      continue;
    }

    if (t.type === 'selfClose') {
      if (t.name === 'br') pushNode(<br key={nextKey()} />);
      continue;
    }

    if (t.type === 'tagOpen') {
      stack.push({
        name: t.name,
        attrs: t.attrs,
        children: [],
        key: nextKey(),
      });
      continue;
    }

    if (t.type === 'tagClose') {
      // pop until matching tag
      let frame: Frame | undefined;

      while (stack.length) {
        const last = stack.pop()!;

        frame = last;
        if (last.name === t.name) break;
      }

      if (!frame) continue;

      const children = frame.children;

      switch (frame.name) {
        case 'b':
          pushNode(<b key={frame.key}>{children}</b>);
          break;
        case 'i':
          pushNode(<i key={frame.key}>{children}</i>);
          break;
        case 'ul':
          pushNode(<ul key={frame.key}>{children}</ul>);
          break;
        case 'li':
          pushNode(<li key={frame.key}>{children}</li>);
          break;
        case 'a': {
          const href = frame.attrs?.href || '#';
          const target = frame.attrs?.target || '_self';
          const rel = frame.attrs?.rel || '';

          pushNode(
            <a key={frame.key} href={href} target={target} rel={rel}>
              {children}
            </a>,
          );
          break;
        }

        default:
          pushNode(<span key={frame.key}>{children}</span>);
      }
    }
  }

  // if tags were not closed properly, flush as text-ish spans
  while (stack.length) {
    const f = stack.pop()!;

    pushNode(<span key={f.key}>{f.children}</span>);
  }

  return out;
}

export function parseHtmlToChildren(html: string): React.ReactNode[] {
  // IMPORTANT: same logic on SSR + client
  const tokens = tokenize(html);

  return buildReactTree(tokens);
}
