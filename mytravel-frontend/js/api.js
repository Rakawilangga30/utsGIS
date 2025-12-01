// Backend base URL - update if necessary
;(function(){
    const BASE = window.API_BASE_OVERRIDE || "https://my-travel-backend-li5d.vercel.app";

    async function apiRequest(path, { method = 'GET', body, headers = {}, credentials = 'include' } = {}) {
        const opts = { method, headers: { ...headers } };
        if (credentials) opts.credentials = credentials;

        if (body instanceof FormData) {
            opts.body = body;
        } else if (body) {
            opts.body = JSON.stringify(body);
            opts.headers['Content-Type'] = 'application/json';
        }

        try {
            const res = await fetch(BASE + path, opts);
            const text = await res.text();
            try { return JSON.parse(text) } catch (e) { return text }
        } catch (err) {
            console.error('API Error:', err);
            return { error: 'Gagal terhubung ke server' };
        }
    }

    // expose globals for older non-module scripts
    window.API = BASE
    window.apiRequest = apiRequest
    // also export for module usage if environment supports it
    try { if(typeof exports !== 'undefined') exports.apiRequest = apiRequest } catch(e){}
})();