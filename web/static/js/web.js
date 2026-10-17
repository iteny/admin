/* Tabler CMS front-end interactions. */
(function () {
  'use strict';

  const themeToggle = document.getElementById('theme-toggle');
  function syncThemeIcon() {
    if (!themeToggle) return;
    const dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
    const icon = themeToggle.querySelector('i');
    if (icon) icon.className = dark ? 'ti ti-sun' : 'ti ti-moon';
  }
  if (themeToggle) {
    syncThemeIcon();
    themeToggle.addEventListener('click', () => {
      const dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
      const next = dark ? 'light' : 'dark';
      document.documentElement.setAttribute('data-bs-theme', next);
      try { localStorage.setItem('tabler-theme', next); } catch (e) { /* ignore */ }
      syncThemeIcon();
    });
  }

  document.addEventListener('keydown', (event) => {
    if (event.key === '/' && !/^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName)) {
      const input = document.querySelector('input[type="search"]');
      if (input) {
        event.preventDefault();
        input.focus();
      }
    }
  });
})();
