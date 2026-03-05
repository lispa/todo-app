// web/app.js (English comments)

const API_URL = '/api'; // Using relative path because Nginx proxies it
let currentAuthMode = 'login'; 

// --- SECTION 1: UI NAVIGATION ---

function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    sections.forEach(s => document.getElementById(s).classList.add('d-none'));
    document.getElementById('nav-user').classList.add('d-none');

    if (section === 'welcome') {
        document.getElementById('welcome-section').classList.remove('d-none');
    } else if (section === 'auth') {
        if (mode) currentAuthMode = mode;
        updateAuthUI();
        document.getElementById('auth-section').classList.remove('d-none');
    } else if (section === 'todo') {
        document.getElementById('todo-section').classList.remove('d-none');
        document.getElementById('nav-user').classList.remove('d-none');
        loadTasks();
    }
}

function updateAuthUI() {
    const signupFields = document.getElementById('signup-fields');
    const title = document.getElementById('auth-title');
    const submitBtn = document.getElementById('auth-submit-btn');
    const switchLink = document.getElementById('auth-switch-link');
    const switchText = document.getElementById('auth-switch-text');

    if (currentAuthMode === 'signup') {
        signupFields.classList.remove('d-none');
        title.innerText = 'Create Account';
        submitBtn.innerText = 'Sign Up';
        switchText.innerText = 'Already have an account? ';
        switchLink.innerText = 'Sign In';
    } else {
        signupFields.classList.add('d-none');
        title.innerText = 'Welcome Back';
        submitBtn.innerText = 'Sign In';
        switchText.innerText = "Don't have an account? ";
        switchLink.innerText = 'Create Account';
    }
}

function toggleAuthMode() {
    currentAuthMode = (currentAuthMode === 'login') ? 'signup' : 'login';
    updateAuthUI();
}

// --- SECTION 2: AUTHENTICATION ---

document.getElementById('auth-form').addEventListener('submit', async (e) => {
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

        if (response.ok) {
            if (currentAuthMode === 'login') {
                localStorage.setItem('token', data.token);
                showSection('todo');
            } else {
                alert('Account created! Now please sign in.');
                currentAuthMode = 'login';
                updateAuthUI();
            }
        } else {
            alert(data.error || 'Authentication failed');
        }
    } catch (err) {
        console.error('Auth error:', err);
        alert('Server connection error');
    }
});

function logout() {
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- SECTION 3: TASK MANAGEMENT ---

async function loadTasks() {
    const token = localStorage.getItem('token');
    const list = document.getElementById('tasks-list');
    
    try {
        const response = await fetch(`${API_URL}/tasks`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.status === 401) return logout();
        
        const tasks = await response.json();
        list.innerHTML = '';
        
        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center animate-fade-in';
            item.innerHTML = `
                <span>${task.title}</span>
                <span class="badge bg-${task.status === 'done' ? 'success' : 'warning'}">${task.status}</span>
            `;
            list.appendChild(item);
        });
    } catch (err) {
        console.error('Load tasks error:', err);
    }
}

async function createTask() {
    const titleInput = document.getElementById('task-title');
    const title = titleInput.value;
    const token = localStorage.getItem('token');

    if (!title) return;

    try {
        const response = await fetch(`${API_URL}/tasks/create`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ title })
        });

        if (response.ok) {
            titleInput.value = '';
            loadTasks();
        }
    } catch (err) {
        console.error('Create task error:', err);
    }
}

// Initialization
window.onload = () => {
    const token = localStorage.getItem('token');
    token ? showSection('todo') : showSection('welcome');
};