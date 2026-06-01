(function () {
	const pageSize = 20;
	const states = new WeakMap();

	function rootFor(el) {
		return el.closest?.("[data-site-browser]") || el;
	}

	function stateFor(root) {
		let state = states.get(root);
		if (!state) {
			state = { page: 1, loading: false, done: false, initialized: false };
			states.set(root, state);
		}
		return state;
	}

	function controls(root) {
		return {
			query: root.querySelector("[name='query']")?.value.trim() || "",
			sort: root.querySelector("[name='sort']")?.value || "top",
		};
	}

	function apiPrefix(root) {
		return (root.dataset.apiPrefix || "").replace(/\/$/, "");
	}

	function endpoint(root, page) {
		const current = controls(root);
		const prefix = apiPrefix(root);
		if (current.query) {
			return `${prefix}/api/public/sites/search/${encodeURIComponent(current.query)}/${page}`;
		}
		return `${prefix}/api/public/sites/${current.sort === "latest" ? "latest" : "top"}/${page}`;
	}

	function status(root, text, kind = "") {
		const target = root.querySelector("[data-site-status]");
		if (target) {
			target.textContent = text;
			target.dataset.status = kind;
		}
	}

	function escapeHTML(value) {
		return String(value ?? "").replace(/[&<>"']/g, (char) => ({
			"&": "&amp;",
			"<": "&lt;",
			">": "&gt;",
			'"': "&quot;",
			"'": "&#39;",
		})[char]);
	}

	function normalizeTag(tag) {
		return String(tag ?? "").replace(/\s+/g, "").toLowerCase();
	}

	function tagColorFor(tag) {
		const normalizedTag = normalizeTag(tag);
		let hash = 0;

		for (let index = 0; index < normalizedTag.length; index += 1) {
			hash = (hash * 31 + normalizedTag.charCodeAt(index)) >>> 0;
		}

		const hue = hash % 360;

		return {
			background: `hsl(${hue}, 62%, 38%)`,
			text: `hsl(${hue}, 72%, 90%)`,
		};
	}

	function renderTag(tag) {
		const color = tagColorFor(tag);
		return `<span class="site-tag" style="background:${color.background};color:${color.text}">#${escapeHTML(tag)}</span>`;
	}

	function renderSite(site) {
		const tags = (site.tags || []).map(renderTag).join("");
		const tagMeta = tags ? `<div class="site-meta">${tags}</div>` : "";
		return `<a class="site-card" href="${escapeHTML(site.url)}">
			<div class="site-card-header">
				<h2>${escapeHTML(site.name)}</h2>
				<span class="site-clicks" aria-label="${escapeHTML(site.clicks)} clicks">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" d="M15.042 21.672 13.684 16.6m0 0-2.51 2.225.569-9.47 5.227 7.917-3.286-.672Zm-7.518-.267A8.25 8.25 0 1 1 20.25 10.5M8.288 14.212A5.25 5.25 0 1 1 17.25 10.5" />
					</svg>
					${escapeHTML(site.clicks)}
				</span>
			</div>
			${tagMeta}
		</a>`;
	}

	async function load(el) {
		const root = rootFor(el);
		const state = stateFor(root);
		if (state.loading || state.done) {
			return;
		}

		state.loading = true;
		status(root, "", "loading");
		try {
			const res = await fetch(endpoint(root, state.page), { headers: { Accept: "application/json" } });
			if (!res.ok) {
				throw new Error(`HTTP ${res.status}`);
			}
			const sites = await res.json();
			const list = root.querySelector("[data-site-list]");
			if (list) {
				list.insertAdjacentHTML("beforeend", sites.map(renderSite).join(""));
			}
			state.page += 1;
			state.done = sites.length < pageSize;
			const empty = state.done && state.page === 2 && sites.length === 0;
			status(root, empty ? "No sites found." : "", empty ? "empty" : "");
		} catch (err) {
			status(root, "Could not load sites.", "empty");
		} finally {
			state.loading = false;
		}
	}

	function reset(el) {
		const root = rootFor(el);
		const state = stateFor(root);
		state.page = 1;
		state.done = false;
		const list = root.querySelector("[data-site-list]");
		if (list) {
			list.textContent = "";
		}
		load(root);
	}

	function init(el) {
		const root = rootFor(el);
		const state = stateFor(root);
		if (state.initialized) {
			return;
		}

		state.initialized = true;
		reset(root);
	}

	function loadIfNearEnd(el) {
		const remaining = document.documentElement.scrollHeight - window.scrollY - window.innerHeight;
		if (remaining < 400) {
			load(el);
		}
	}

	function initAll() {
		document.querySelectorAll("[data-site-browser]").forEach(init);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initAll, { once: true });
	} else {
		initAll();
	}

	window.ConexSites = { init, reset, loadIfNearEnd };
})();
