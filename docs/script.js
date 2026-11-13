const links = [...document.querySelectorAll('nav a')];
const menu = document.querySelector('#menu');
function closeMenu() { document.body.classList.remove('menu-open'); menu.setAttribute('aria-expanded', 'false'); }
menu.addEventListener('click', () => {
 const open = document.body.classList.toggle('menu-open');
 menu.setAttribute('aria-expanded', String(open));
});
document.addEventListener('keydown', event => { if (event.key === 'Escape') { closeMenu(); menu.focus(); } });
links.forEach(link => link.addEventListener('click', () => { closeMenu(); document.querySelector(link.hash)?.scrollIntoView(); }));
document.querySelector('#search').addEventListener('input', event => {
 const query = event.target.value.trim().toLowerCase();
 links.forEach(link => { link.hidden = !link.textContent.toLowerCase().includes(query); });
 document.querySelector('#empty').hidden = links.some(link => !link.hidden);
});
function markActive() {
 let active = links[0];
 links.forEach(link => { if (document.querySelector(link.hash).getBoundingClientRect().top <= 130) active = link; });
 links.forEach(link => {
  link.classList.toggle('active', link === active);
  if (link === active) link.setAttribute('aria-current', 'location'); else link.removeAttribute('aria-current');
 });
}
window.addEventListener('scroll', markActive, {passive:true});
markActive();
document.querySelectorAll('.copy').forEach(button => button.addEventListener('click', async () => {
 const code = button.closest('.code-block').querySelector('code');
 try {
  await navigator.clipboard.writeText(code.textContent);
  button.textContent = 'Copied';
  document.querySelector('#announce').textContent = 'Code copied to clipboard.';
 } catch {
  const range = document.createRange(); range.selectNodeContents(code);
  const selection = window.getSelection(); selection.removeAllRanges(); selection.addRange(range);
  button.textContent = 'Selected';
  document.querySelector('#announce').textContent = 'Code selected. Use your keyboard copy shortcut.';
 }
 setTimeout(() => { button.textContent = 'Copy'; }, 2000);
}));
// Lightweight Go highlighting; textContent is preserved for copying.
document.querySelectorAll('.code-block').forEach(block => {
 if (block.querySelector('.code-toolbar span').textContent !== 'go') return;
 const code = block.querySelector('code');
 const source = code.textContent;
 const pattern = /\/\/[^\n]*|"(?:\\.|[^"\\])*"|\b(?:package|import|func|type|struct|interface|return|if|nil|true|false|var)\b/g;
 const fragment = document.createDocumentFragment(); let cursor = 0;
 for (const match of source.matchAll(pattern)) {
  fragment.append(document.createTextNode(source.slice(cursor, match.index)));
  const span = document.createElement('span'); span.textContent = match[0];
  span.className = match[0].startsWith('//') ? 'tok-comment' : match[0].startsWith('"') ? 'tok-string' : 'tok-keyword';
  fragment.append(span); cursor = match.index + match[0].length;
 }
 fragment.append(document.createTextNode(source.slice(cursor))); code.replaceChildren(fragment);
});
