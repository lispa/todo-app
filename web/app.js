/**
 * Todo-App Frontend Logic
 * Features: SPA Navigation, Auth (Login/Signup), Task Management
 * Comments: English
 */

const API_URL = '/api';

// --- 1. UI NAVIGATION CONTROLLER ---

function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    
    // Hide all sections
    sections.forEach(s => {
        const el = document.getElementById(s);
        if (el) el.classList.add('d-none');
    });

    // Reset Navigation UI
    document.getElementById('nav-user').classList.add('d-none');

    if (section === 'welcome') {
        document.getElementById('welcome-section').classList.remove('d-none');
    } 
    else if (section === 'auth') {
        if (mode) currentAuthMode = mode;
        updateAuthUI();
        document.getElementById('auth-section').classList.remove('d-none');
    } 
    else if (section === 'todo') {
        const token = localStorage.getItem('token');
        if (!token) {
            showSection('welcome');
            return;
        }
        document.getElementById('todo-section').classList.remove('d-none');
        document.getElementById('nav-user').classList.remove('d-none');
        loadTasks(); // Fetch data from server
    }
}

// Global state for Auth Mode
let currentAuthMode = 'login'; 

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

// --- 2. AUTHENTICATION LOGIC ---

document.getElementById('auth-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    console.log(`Attempting ${currentAuthMode}...`);

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
        console.log("Server response:", data);

        if (response.ok) {
            if (currentAuthMode === 'login') {
                // IMPORTANT: Ensure your Go backend returns field named "token"
                if (data.token) {
                    localStorage.setItem('token', data.token);
                    console.log("Token saved. Redirecting to Todo list.");
                    showSection('todo');
                } else {
                    alert("Error: No token received from server.");
                }
            } else {
                alert('Success! Now you can sign in.');
                currentAuthMode = 'login';
                updateAuthUI();
            }
        } else {
            alert(data.error || 'Authentication failed');
        }
    } catch (err) {
        console.error('Auth error:', err);
        alert('Server unreachable');
    }
});

function logout() {
    console.log("Logging out...");
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- 3. TASK MANAGEMENT ---

async function loadTasks() {
    const token = localStorage.getItem('token');
    const list = document.getElementById('tasks-list');
    
    if (!token) return logout();

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            method: 'GET',
            headers: { 
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.status === 401) {
            console.warn("Unauthorized access - logging out");
            return logout();
        }
        
        const tasks = await response.json();
        list.innerHTML = '';
        
        if (tasks.length === 0) {
            list.innerHTML = '<div class="text-center p-4 text-muted">No tasks found. Add your first one!</div>';
            return;
        }

        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center animate-fade-in shadow-sm mb-2 border-0 rounded';
            item.innerHTML = `
                <div>
                    <h6 class="mb-0">${task.title}</h6>
                    <small class="text-muted">Status: ${task.status}</small>
                </div>
                <span class="badge rounded-pill bg-${task.status === 'done' ? 'success' : 'primary'}">
                    ${task.status}
                </span>
            `;
            list.appendChild(item);
        });
    } catch (err) {
        console.error('Failed to load tasks:', err);
    }
}

async function createTask() {
    const titleInput = document.getElementById('task-title');
    const title = titleInput.value.trim();
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
            loadTasks(); // Refresh list
        } else {
            const errorData = await response.json();
            alert(errorData.error || 'Failed to create task');
        }
    } catch (err) {
        console.error('Create task error:', err);
    }
}

// --- 4. INITIALIZATION ---

window.onload = () => {
    const token = localStorage.getItem('token');
    console.log("App initialized. Token present:", !!token);
    
    if (token) {
        showSection('todo');
    } else {
        showSection('welcome');
    }
};