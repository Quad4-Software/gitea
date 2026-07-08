import {GET} from '../../modules/fetch.ts';

const {appSubUrl} = window.config;

export function initAdminReticulumConsole() {
  const consoleEl = document.querySelector<HTMLPreElement>('#reticulum-console');
  if (!consoleEl) return;

  let since = 0;
  let polling = false;

  const poll = async () => {
    if (polling) return;
    polling = true;
    try {
      const resp = await GET(`${appSubUrl}/-/admin/reticulum/logs?since=${since}`);
      const json: {lines?: string[]; next?: number} = await resp.json();
      if (json.lines?.length) {
        const atBottom = consoleEl.scrollTop + consoleEl.clientHeight >= consoleEl.scrollHeight - 8;
        consoleEl.textContent += `${json.lines.join('\n')}\n`;
        since = json.next ?? since;
        if (atBottom) {
          consoleEl.scrollTop = consoleEl.scrollHeight;
        }
      } else if (json.next != null) {
        since = json.next;
      }
    } finally {
      polling = false;
    }
  };

  void poll();
  window.setInterval(() => void poll(), 2000);
}
