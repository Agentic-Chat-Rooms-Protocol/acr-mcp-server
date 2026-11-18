import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');

test('package.json uses @acr-js/mcp-server at version 1.1.0', () => {
  const pkg = JSON.parse(fs.readFileSync(path.join(rootDir, 'package.json'), 'utf8'));
  assert.equal(pkg.name, '@acr-js/mcp-server');
  assert.equal(pkg.version, '1.1.0');
});

test('dist/index.js responds to MCP initialize and tools/list over stdio', async () => {
  const entry = path.join(rootDir, 'dist', 'index.js');
  assert.equal(fs.existsSync(entry), true, 'dist/index.js must exist after build');

  const child = spawn(process.execPath, [entry], { stdio: ['pipe', 'pipe', 'pipe'] });
  let stdout = '';
  let stderr = '';
  child.stdout.on('data', (chunk) => {
    stdout += chunk.toString('utf8');
  });
  child.stderr.on('data', (chunk) => {
    stderr += chunk.toString('utf8');
  });

  const initReq = JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'initialize',
    params: {
      protocolVersion: '2024-11-05',
      capabilities: {},
      clientInfo: { name: 'test-client', version: '1.0.0' },
    },
  });
  const listReq = JSON.stringify({
    jsonrpc: '2.0',
    id: 2,
    method: 'tools/list',
    params: {},
  });

  child.stdin.write(initReq + '\n');
  child.stdin.write(listReq + '\n');
  child.stdin.end();

  await new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      child.kill();
      reject(new Error(`timeout waiting for mcp-server stdout: ${stdout} stderr: ${stderr}`));
    }, 5000);
    child.on('exit', () => {
      clearTimeout(timer);
      resolve();
    });
  });

  const lines = stdout
    .trim()
    .split(/\r?\n/)
    .filter(Boolean)
    .map((l) => JSON.parse(l));

  assert.equal(lines.length, 2, `expected 2 JSON-RPC responses, got ${lines.length} (stderr=${stderr})`);
  assert.equal(lines[0].result.serverInfo.name, 'acr-mcp-server');
  assert.equal(lines[0].result.serverInfo.version, '1.1.0');
  const toolNames = lines[1].result.tools.map((t) => t.name);
  assert.deepEqual(toolNames, ['chat_send_message', 'chat_cast_vote', 'mcp_ast_diff_verify']);
});
