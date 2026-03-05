/**
 * Todo-App Final Frontend Logic
 */

const API_URL = '/api';
let currentAuthMode = 'login'; 

// --- 1. UI NAVIGATION ---
function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    sections.forEach(s => document.getElementById(s)?.classList.add('d-none'));
    document.getElementById('nav-user')?.classList.add('d-none');

    if (section === 'welcome') {
        document.getElementById('welcome-section')?.classList.remove('d-none');
    } else if (section === 'auth') {
        if (mode) currentAuthMode = mode;
        updateAuthUI();
        document.getElementById('auth-section')?.classList.remove('d-none');
    } else if (section === 'todo') {
        if (!localStorage.getItem('token')) return showSection('welcome');
        document.getElementById('todo-section')?.classList.remove('d-none');
        document.getElementById('nav-user')?.classList.remove('d-none');
        loadTasks();
    }
}

function updateAuthUI() {
    const signupFields = document.getElementById('signup-fields');
    const title = document.getElementById('auth-title');
    const submitBtn = document.getElementById('auth-submit-btn');
    const switchLink = document.getElementById('auth-switch-link');

    if (currentAuthMode === 'signup') {
        signupFields?.classList.remove('d-none');
        if (title) title.innerText = 'Create Account';
        if (submitBtn) submitBtn.innerText = 'Sign Up';
        if (switchLink) switchLink.innerText = 'Sign In';
    } else {
        signupFields?.classList.add('d-none');
        if (title) title.innerText = 'Welcome Back';
        if (submitBtn) submitBtn.innerText = 'Sign In';
        if (switchLink) switchLink.innerText = 'Create Account';
    }
}

function toggleAuthMode() {
    currentAuthMode = (currentAuthMode === 'login') ? 'signup' : 'login';
    updateAuthUI();
}

// --- 2. AUTHENTICATION ---
document.getElementById('auth-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const endpoint = currentAuthMode === 'signup' ? '/auth/signup' : '/auth/login';
    
    const payload = { email, password };
    if (currentAuthMode === 'signup') {
        payload.first_name = document.getElementById('first_name').value;
        payload.last_name = document.getElementById('last_name').value;
    }

    try {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const data = await response.json();
        console.log("Full Server Response:", data); // СМОТРИМ СЮДА В КОНСОЛИ

        if (response.ok) {
            if (currentAuthMode === 'login') {
                // Пытаемся найти токен в разных полях (универсальный поиск)
                const token = data.token || data.access_token || data.jwt;
                
                if (token) {
                    localStorage.setItem('token', token);
                    console.log("Token captured and saved!");
                    showSection('todo');
                } else {
                    console.error("Token field not found in JSON!", data);
                    alert("Auth success, but token is missing in response.");
                }
            } else {
                alert('Success! Now Sign In.');
                toggleAuthMode();
            }
        } else {
            alert(data.error || 'Check your credentials');
        }
    } catch (err) {
        console.error('Network error:', err);
    }
});

function logout() {
    console.warn("Logging out...");
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- 3. TASKS ---
async function loadTasks() {
    const token = localStorage.getItem('token');
    if (!token) return logout();

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.status === 401) return logout();

        const tasks = await response.json();
        const list = document.getElementById('tasks-list');
        if (!list) return;

        list.innerHTML = tasks.length === 0 
            ? '<div class="text-center p-4">No tasks yet!</div>' 
            : '';

        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center shadow-sm mb-2 rounded border-0 animate-fade-in';
            item.innerHTML = `
                <div>
                    <h6 class="mb-0">${task.title}</h6>
                    <small class="text-muted">${task.status}</small>
                </div>
                <span class="badge bg-primary rounded-pill">${task.status}</span>
            `;
            list.appendChild(item);
        });
    } catch (err) {
        console.error('Load tasks error:', err);
    }
}

async function createTask() {
    const titleInput = document.getElementById('task-title');
    const token = localStorage.getItem('token');
    if (!titleInput.value || !token) return;

    try {
        const response = await fetch(`${API_URL}/tasks/create`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ title: titleInput.value })
        });

        if (response.ok) {
            titleInput.value = '';
            loadTasks();
        }
    } catch (err) { console.error(err); }
}

// --- 4. INIT ---
window.onload = () => {
    const token = localStorage.getItem('token');
    token ? showSection('todo') : showSection('welcome');
};