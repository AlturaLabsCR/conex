(function () {
	const pageSize = 20;
	const states = new WeakMap();

	function rootFor(el) {
		return el.closest?.("[data-site-browser]") || el;
	}

	function stateFor(root) {
		let state = states.get(root);
		if (!state) {
			state = { page: 1, loading: false, done: false };
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

	function status(root, text) {
		const target = root.querySelector("[data-site-status]");
		if (target) {
			target.textContent = text;
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

	function renderSite(site) {
		const tags = (site.tags || []).map((tag) => `#${escapeHTML(tag)}`).join(" ");
		return `<article class="site-card">
			<h2><a href="${escapeHTML(site.url)}">${escapeHTML(site.name)}</a></h2>
			<div class="site-meta">${escapeHTML(site.clicks)} clicks · ${tags}</div>
		</article>`;
	}

	async function load(el) {
		const root = rootFor(el);
		const state = stateFor(root);
		if (state.loading || state.done) {
			return;
		}

		state.loading = true;
		status(root, "Loading...");
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
			status(root, state.done ? "No more sites." : "");
		} catch (err) {
			status(root, "Could not load sites.");
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

	function loadIfNearEnd(el) {
		const remaining = document.documentElement.scrollHeight - window.scrollY - window.innerHeight;
		if (remaining < 400) {
			load(el);
		}
	}

	window.ConexSites = { init: reset, reset, loadIfNearEnd };
})();
