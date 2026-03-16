/**
 * Todo-App Frontend Logic
 * Handles navigation, auth, and dynamic task rendering.
 * @version 1.4
 */

const API_URL = '/api';
let currentAuthMode = 'login'; 

// --- SECTION 1: NAVIGATION ---

/**
 * Switch visibility between app sections.
 */
function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    
    sections.forEach(s => {
        const el = document.getElementById(s);
        if (el) el.classList.add('d-none');
    });

    document.getElementById('nav-user')?.classList.add('d-none');

    if (section === 'welcome') {
        document.getElementById('welcome-section')?.classList.remove('d-none');
    } 
    else if (section === 'auth') {
        if (mode) currentAuthMode = mode;
        updateAuthUI();
        document.getElementById('auth-section')?.classList.remove('d-none');
    } 
    else if (section === 'todo') {
        const token = localStorage.getItem('token');
        if (!token) return showSection('welcome');
        
        document.getElementById('todo-section')?.classList.remove('d-none');
        document.getElementById('nav-user')?.classList.remove('d-none');
        loadTasks();
    }
}

/**
 * Toggle between Login and Signup text in UI.
 */
function updateAuthUI() {
    const signupFields = document.getElementById('signup-fields');
    const title = document.getElementById('auth-title');
    const submitBtn = document.getElementById('auth-submit-btn');
    const switchLink = document.getElementById('auth-switch-link');

    if (currentAuthMode === 'signup') {
        signupFields?.classList.remove('d-none');
        if (title) title.innerText = 'Create Your Account';
        if (submitBtn) submitBtn.innerText = 'Register';
        if (switchLink) switchLink.innerText = 'Already have an account? Sign In';
    } else {
        signupFields?.classList.add('d-none');
        if (title) title.innerText = 'Welcome Back';
        if (submitBtn) submitBtn.innerText = 'Sign In';
        if (switchLink) switchLink.innerText = 'New here? Create an Account';
    }
}

function toggleAuthMode() {
    currentAuthMode = (currentAuthMode === 'login') ? 'signup' : 'login';
    updateAuthUI();
}

// --- SECTION 2: AUTHENTICATION ---

document.getElementById('auth-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const endpoint = currentAuthMode === 'signup' ? '/auth/signup' : '/auth/login';
    
    const payload = { email, password };
    if (currentAuthMode === 'signup') {
        payload.first_name = document.getElementById('first_name').value || "Guest";
        payload.last_name = document.getElementById('last_name').value || "User";
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
                const token = data.token || data.access_token;
                if (token) {
                    localStorage.setItem('token', token);
                    showSection('todo');
                }
            } else {
                alert('Account created! Please sign in.');
                currentAuthMode = 'login';
                updateAuthUI();
            }
        } else {
            alert(data.error || 'Auth failed');
        }
    } catch (err) { console.error('Auth Error:', err); }
});

function logout() {
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- SECTION 3: TASK MANAGEMENT ---

/**
 * Fetch tasks and render them with action buttons (Start/Done/Delete).
 */
async function loadTasks() {
    const token = localStorage.getItem('token');
    if (!token) return showSection('welcome');

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            method: 'GET',
            headers: { 
                'Authorization': `Bearer ${token.trim()}`,
                'Accept': 'application/json'
            }
        });

        if (response.status === 401) return logout();

        const tasks = await response.json();
        const list = document.getElementById('tasks-list');
        if (!list) return;

        list.innerHTML = tasks.length === 0 ? '<div class="text-center p-3 text-muted">No tasks found.</div>' : '';

        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center shadow-sm mb-3 border-0 rounded p-3';
            
            // Task status logic
            const isDone = task.status === 'done';
            const isInProgress = task.status === 'in_progress';
            const badgeClass = isDone ? 'bg-success' : (isInProgress ? 'bg-warning text-dark' : 'bg-primary');

            // Render task with conditional buttons
            item.innerHTML = `
                <div class="flex-grow-1">
                    <h6 class="mb-1 fw-bold ${isDone ? 'text-decoration-line-through text-muted' : ''}">${task.title}</h6>
                    <span class="badge ${badgeClass}">${task.status.replace('_', ' ')}</span>
                </div>
                <div class="btn-group ms-3">
                    ${task.status === 'todo' ? 
                        `<button class="btn btn-sm btn-outline-warning" onclick="updateTaskStatus(${task.id}, 'start')">Start</button>` : ''}
                    
                    ${isInProgress ? 
                        `<button class="btn btn-sm btn-outline-success" onclick="updateTaskStatus(${task.id}, 'done')">Done</button>` : ''}
                    
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteTask(${task.id})">Delete</button>
                </div>
            `;
            list.appendChild(item);
        });
    } catch (err) { console.error('Load Error:', err); }
}

/**
 * Generic status update (Start/Done).
 */
async function updateTaskStatus(taskId, action) {
    const token = localStorage.getItem('token');
    try {
        const response = await fetch(`${API_URL}/tasks/${action}`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token.trim()}`
            },
            body: JSON.stringify({ id: taskId })
        });

        if (response.ok) loadTasks();
    } catch (err) { console.error(`Error during ${action}:`, err); }
}

/**
 * Remove task from server.
 */
async function deleteTask(taskId) {
    if (!confirm("Are you sure?")) return;
    const token = localStorage.getItem('token');
    try {
        const response = await fetch(`${API_URL}/tasks/delete`, {
            method: 'DELETE',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token.trim()}`
            },
            body: JSON.stringify({ id: taskId })
        });

        if (response.ok) loadTasks();
    } catch (err) { console.error('Delete error:', err); }
}

// Ensure functions are global for HTML onclicks
window.updateTaskStatus = updateTaskStatus;
window.deleteTask = deleteTask;

// --- SECTION 4: INITIALIZATION ---

window.onload = () => {
    const token = localStorage.getItem('token');
    token ? showSection('todo') : showSection('welcome');
};