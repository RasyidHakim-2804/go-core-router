// Apply before CSS loads to avoid a flash of the wrong theme.
(() => {
 const key = 'go-core-router-theme';
 const system = window.matchMedia('(prefers-color-scheme: dark)');
 let preference = null;
 try {
  const saved = localStorage.getItem(key);
  if (saved === 'dark' || saved === 'light') preference = saved;
 } catch { /* Storage may be unavailable; keep the choice in memory. */ }
 function apply(theme) {
  document.documentElement.dataset.theme = theme;
  const button = document.querySelector('#theme-toggle');
  if (button) {
   button.setAttribute('aria-pressed', String(theme === 'dark'));
   button.textContent = theme === 'dark' ? '☀ Light mode' : '☾ Dark mode';
  }
 }
 const current = () => preference || (system.matches ? 'dark' : 'light');
 apply(current());
 system.addEventListener('change', () => { if (!preference) apply(current()); });
 window.addEventListener('storage', event => {
  if (event.key !== key && event.key !== null) return;
  preference = event.newValue === 'dark' || event.newValue === 'light' ? event.newValue : null;
  apply(current());
 });
 document.addEventListener('DOMContentLoaded', () => {
  apply(current());
  document.querySelector('#theme-toggle').addEventListener('click', () => {
   preference = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
   try { localStorage.setItem(key, preference); } catch { /* Session-only fallback. */ }
   apply(preference);
  });
 });
})();
