/* Tabler CMS admin interactions. No framework, no build step. */
(function () {
  'use strict';

  const bootstrap = window.bootstrap || (window.tabler && window.tabler.bootstrap);
  const csrf = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '';

  function toast(message, type) {
    if (!bootstrap) return;
    const wrap = document.getElementById('cms-toast-wrap') || (() => {
      const el = document.createElement('div');
      el.id = 'cms-toast-wrap';
      el.className = 'toast-container position-fixed top-0 end-0 p-3';
      el.style.zIndex = '1090';
      document.body.appendChild(el);
      return el;
    })();
    const el = document.createElement('div');
    el.className = 'toast align-items-center text-bg-' + (type || 'success') + ' border-0';
    el.setAttribute('role', 'alert');
    el.innerHTML = '<div class="d-flex"><div class="toast-body"></div><button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="关闭"></button></div>';
    el.querySelector('.toast-body').textContent = message;
    wrap.appendChild(el);
    const instance = new bootstrap.Toast(el, { delay: 3000 });
    el.addEventListener('hidden.bs.toast', () => el.remove());
    instance.show();
  }

  function post(url, data) {
    const body = data instanceof URLSearchParams ? data : new URLSearchParams(data || {});
    return fetch(url, {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrf, 'X-Requested-With': 'XMLHttpRequest' },
      body,
    }).then(async (res) => {
      const json = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(json.msg || '请求失败');
      return json;
    });
  }

  // Theme toggle
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

  // Delete confirmation
  const confirmModalEl = document.getElementById('modal-confirm');
  let confirmModal = null;
  document.addEventListener('click', (event) => {
    const trigger = event.target.closest('.confirm-delete');
    if (!trigger || !confirmModalEl) return;
    event.preventDefault();
    const form = document.getElementById('confirm-form');
    const message = document.getElementById('confirm-message');
    if (form) form.setAttribute('action', trigger.dataset.url || '#');
    if (message) message.textContent = '将删除「' + (trigger.dataset.name || '此项') + '」，此操作不可撤销。';
    confirmModal = confirmModal || new bootstrap.Modal(confirmModalEl);
    confirmModal.show();
  });

  // Copy URL
  document.addEventListener('click', async (event) => {
    const trigger = event.target.closest('.copy-url');
    if (!trigger) return;
    event.preventDefault();
    const url = trigger.dataset.url || '';
    try {
      await navigator.clipboard.writeText(new URL(url, window.location.origin).href);
      toast('URL 已复制');
    } catch (e) {
      toast('复制失败，请手动复制：' + url, 'danger');
    }
  });

  // Media picker modal (iframe)
  let mediaModal = null;
  let mediaTarget = null;
  let hugerteCallback = null;

  function openMediaPicker(target) {
    mediaTarget = target;
    if (!mediaModal) {
      const el = document.createElement('div');
      el.className = 'modal modal-blur fade';
      el.id = 'cms-media-modal';
      el.tabIndex = -1;
      el.innerHTML = '<div class="modal-dialog modal-xl modal-dialog-centered"><div class="modal-content"><div class="modal-header"><h5 class="modal-title">选择媒体</h5><button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="关闭"></button></div><div class="modal-body p-0"><iframe src="/admin/media/picker" title="媒体库" style="width:100%;height:70vh;border:0;display:block"></iframe></div></div></div>';
      document.body.appendChild(el);
      mediaModal = new bootstrap.Modal(el);
      el.addEventListener('hidden.bs.modal', () => {
        hugerteCallback = null;
        mediaTarget = null;
      });
    } else {
      const iframe = mediaModal._element.querySelector('iframe');
      if (iframe) iframe.src = '/admin/media/picker';
    }
    mediaModal.show();
  }

  document.addEventListener('click', (event) => {
    const trigger = event.target.closest('.media-picker');
    if (!trigger) return;
    event.preventDefault();
    openMediaPicker(trigger.dataset.target);
  });

  window.addEventListener('message', (event) => {
    if (event.origin !== window.location.origin) return;
    const data = event.data || {};
    if (data.type !== 'media-selected') return;
    if (mediaTarget === '__hugerte' && hugerteCallback) {
      hugerteCallback({ files: [data.url] });
      hugerteCallback = null;
    } else if (mediaTarget) {
      const input = document.getElementById(mediaTarget);
      if (input) {
        input.value = data.url;
        input.dispatchEvent(new Event('change'));
      }
    }
    mediaTarget = null;
    if (mediaModal) mediaModal.hide();
  });

  // Picker page: selecting an item posts back to the opener.
  document.addEventListener('click', (event) => {
    const item = event.target.closest('.media-picker-item');
    if (!item) return;
    event.preventDefault();
    const payload = { type: 'media-selected', url: item.dataset.url, id: item.dataset.id };
    if (window.parent && window.parent !== window) {
      window.parent.postMessage(payload, window.location.origin);
    } else if (window.opener) {
      window.opener.postMessage(payload, window.location.origin);
    }
  });

  // Dropzone
  if (window.Dropzone) {
    window.Dropzone.autoDiscover = false;
    const dz = document.getElementById('media-dropzone');
    if (dz) {
      const max = Number(dz.dataset.maxSize || 10 * 1024 * 1024);
      new Dropzone(dz, {
        url: '/admin/media/upload',
        headers: { 'X-CSRF-Token': csrf, 'X-Requested-With': 'XMLHttpRequest' },
        maxFilesize: Math.max(1, Math.round(max / 1024 / 1024)),
        acceptedFiles: 'image/*',
        addRemoveLinks: false,
        success: function () { window.location.reload(); },
        error: function (file, response) {
          let msg = '上传失败';
          if (typeof response === 'string') {
            msg = response;
          } else if (response) {
            if (response.msg) msg = response.msg;
            else if (response.responseText) {
              try { msg = JSON.parse(response.responseText).msg || msg; } catch (e) {}
            }
          }
          toast(msg, 'danger');
          this.removeFile(file);
        },
      });
    }
  }

  // Paste image into media library
  if (window.location.pathname === '/admin/media') {
    document.addEventListener('paste', async (event) => {
      const items = event.clipboardData && event.clipboardData.items;
      if (!items) return;
      for (const item of items) {
        if (!item.type || item.type.indexOf('image/') !== 0) continue;
        const file = item.getAsFile();
        if (!file) continue;
        const data = new FormData();
        data.append('file', file, 'pasted-' + Date.now() + '.png');
        try {
          const res = await fetch('/admin/media/upload', {
            method: 'POST',
            headers: { 'X-CSRF-Token': csrf, 'X-Requested-With': 'XMLHttpRequest' },
            body: data,
          });
          const json = await res.json();
          if (!res.ok || json.code !== 0) throw new Error(json.msg || '上传失败');
          toast('图片已上传');
          window.location.reload();
        } catch (e) {
          toast(e.message, 'danger');
        }
        event.preventDefault();
        break;
      }
    });
  }

  // Sortable lists
  if (window.Sortable) {
    const lists = [
      { el: document.getElementById('category-tree'), url: '/admin/categories/reorder' },
      { el: document.getElementById('menu-tree'), url: '/admin/menus/reorder' },
    ];
    lists.forEach(({ el, url }) => {
      if (!el) return;
      new Sortable(el, {
        handle: '.drag-handle',
        animation: 150,
        onEnd: async function () {
          const ids = Array.from(el.querySelectorAll('tr[data-id]')).map((row) => row.dataset.id);
          const body = new URLSearchParams();
          ids.forEach((id) => body.append('ids[]', id));
          try {
            await post(url, body);
            toast('排序已保存');
            window.location.reload();
          } catch (e) {
            toast(e.message, 'danger');
          }
        },
      });
    });
  }

  // Tom Select
  if (window.TomSelect) {
    document.querySelectorAll('select.tom-select').forEach((el) => {
      try {
        new TomSelect(el, { maxItems: null });
      } catch (e) {
        console.error('Tom Select init failed:', e);
      }
    });
  }

  // HugeRTE
  if (window.hugerte) {
    document.querySelectorAll('textarea.rich-editor').forEach((el) => {
      try {
        hugerte.init({
          target: el,
          menubar: false,
          branding: false,
          promotion: false,
          height: 460,
          plugins: 'lists link image table code autoresize',
          toolbar: 'undo redo | blocks | bold italic underline | bullist numlist | blockquote link cmsimage table | code',
          content_style: 'img{max-width:100%;height:auto}',
          convert_urls: false,
          relative_urls: false,
          remove_script_host: false,
          image_title: true,
          automatic_uploads: false,
          file_picker_types: 'image',
          file_picker: function (callback) {
            hugerteCallback = callback;
            openMediaPicker('__hugerte');
          },
          setup: function (editor) {
            editor.on('change', function () { editor.save(); });
            editor.ui.registry.addButton('cmsimage', {
              icon: 'image',
              tooltip: '从媒体库插入图片',
              onAction: function () {
                hugerteCallback = function (result) {
                  const url = Array.isArray(result.files) ? result.files[0] : result;
                  if (url) {
                    editor.insertContent('<img src="' + String(url).replace(/"/g, '&quot;') + '" alt="" />');
                  }
                };
                openMediaPicker('__hugerte');
              },
            });
          },
        });
      } catch (e) {
        console.error('HugeRTE init failed:', e);
      }
    });
  }

  // Batch checkbox
  const checkAll = document.getElementById('check-all');
  if (checkAll) {
    checkAll.addEventListener('change', () => {
      document.querySelectorAll('.row-check').forEach((el) => { el.checked = checkAll.checked; });
    });
  }

  // Permission select all
  const permAll = document.getElementById('perm-check-all');
  if (permAll) {
    permAll.addEventListener('change', () => {
      document.querySelectorAll('.perm-check').forEach((el) => { el.checked = permAll.checked; });
    });
  }

  // Auto slug
  const title = document.getElementById('title');
  const slug = document.getElementById('slug');
  if (title && slug) {
    let manual = slug.value.trim() !== '';
    slug.addEventListener('input', () => { manual = true; });
    title.addEventListener('input', () => {
      if (manual) return;
      const generated = title.value.toLowerCase().replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-').replace(/^-+|-+$/g, '');
      slug.value = generated.replace(/[\u4e00-\u9fa5]/g, '');
      const preview = document.getElementById('slug-preview');
      if (preview) preview.textContent = slug.value;
    });
  }

  // Category modal
  window.resetCategoryForm = function () {
    const form = document.getElementById('category-form');
    if (!form) return;
    form.reset();
    form.action = '/admin/categories';
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = '';
    const title = document.getElementById('category-modal-title');
    if (title) title.textContent = '新增分类';
    const parent = document.getElementById('category-parent');
    if (parent) parent.value = '0';
    const status = document.getElementById('category-status');
    if (status) status.value = '1';
  };

  window.editCategory = function (button) {
    const row = button.closest('tr');
    const form = document.getElementById('category-form');
    if (!row || !form) return;
    const set = (name, value) => { const el = form.querySelector('[name="' + name + '"]'); if (el) el.value = value || ''; };
    form.action = '/admin/categories/' + row.dataset.id;
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = 'PUT';
    set('name', row.dataset.name);
    set('slug', row.dataset.slug);
    set('parent_id', row.dataset.parent);
    set('sort', row.dataset.sort);
    set('status', row.dataset.status);
    set('description', row.dataset.description);
    set('cover_url', row.dataset.cover);
    set('seo_title', row.dataset.seoTitle);
    set('seo_keywords', row.dataset.seoKeywords);
    set('seo_description', row.dataset.seoDescription);
    const title = document.getElementById('category-modal-title');
    if (title) title.textContent = '编辑分类';
    new bootstrap.Modal(document.getElementById('modal-category')).show();
  };

  // Tag modal
  window.resetTagForm = function () {
    const form = document.getElementById('tag-form');
    if (!form) return;
    form.reset();
    form.action = '/admin/tags';
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = '';
    const title = document.getElementById('tag-modal-title');
    if (title) title.textContent = '新增标签';
  };

  window.editTag = function (button) {
    const row = button.closest('tr');
    const form = document.getElementById('tag-form');
    if (!row || !form) return;
    form.action = '/admin/tags/' + row.dataset.id;
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = 'PUT';
    document.getElementById('tag-name').value = row.dataset.name || '';
    document.getElementById('tag-slug').value = row.dataset.slug || '';
    const title = document.getElementById('tag-modal-title');
    if (title) title.textContent = '编辑标签';
    new bootstrap.Modal(document.getElementById('modal-tag')).show();
  };

  // Menu modal
  window.resetMenuForm = function () {
    const form = document.getElementById('menu-form');
    if (!form) return;
    form.reset();
    form.action = '/admin/menus';
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = '';
    const title = document.getElementById('menu-modal-title');
    if (title) title.textContent = '新增菜单';
    const parent = document.getElementById('menu-parent');
    if (parent) parent.value = '0';
    const visible = document.getElementById('menu-visible');
    if (visible) visible.checked = true;
  };

  window.editMenu = function (button) {
    const row = button.closest('tr');
    const form = document.getElementById('menu-form');
    if (!row || !form) return;
    form.action = '/admin/menus/' + row.dataset.id;
    const method = form.querySelector('input[name="_method"]');
    if (method) method.value = 'PUT';
    const set = (name, value) => { const el = form.querySelector('[name="' + name + '"]'); if (el) el.value = value || ''; };
    set('name', row.dataset.name);
    set('icon', row.dataset.icon);
    set('parent_id', row.dataset.parent);
    set('path', row.dataset.path);
    set('permission_code', row.dataset.permission);
    set('sort', row.dataset.sort);
    const visible = document.getElementById('menu-visible');
    if (visible) visible.checked = row.dataset.visible === 'true';
    const external = document.getElementById('menu-external');
    if (external) external.checked = row.dataset.external === 'true';
    const title = document.getElementById('menu-modal-title');
    if (title) title.textContent = '编辑菜单';
    new bootstrap.Modal(document.getElementById('modal-menu')).show();
  };

  // Dashboard chart
  const trendEl = document.getElementById('trend-data');
  const chartEl = document.getElementById('chart-trend');
  if (trendEl && chartEl && window.ApexCharts) {
    try {
      const trend = JSON.parse(trendEl.textContent || '[]');
      new ApexCharts(chartEl, {
        chart: { type: 'area', height: 260, toolbar: { show: false } },
        series: [{ name: '发布数', data: trend.map((item) => item.count) }],
        xaxis: { categories: trend.map((item) => item.date) },
        dataLabels: { enabled: false },
        stroke: { curve: 'smooth', width: 2 },
        colors: ['#066fd1'],
        fill: { type: 'gradient', gradient: { opacityFrom: 0.35, opacityTo: 0.05 } },
        grid: { strokeDashArray: 4 },
      }).render();
    } catch (e) {
      console.error(e);
    }
  }
})();
