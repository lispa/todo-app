/**
 * Todo-App Frontend Logic
 * Refactored version with API Wrapper and clean separation of concerns.
 * @version 1.5
 */

// --- 1. API CONFIGURATION & WRAPPER ---

const API = {
    baseUrl: '/api',

    // Get current auth headers
    getHeaders() {
        const token = localStorage.getItem('token');
        return {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            ...(token ? { 'Authorization': `Bearer ${token.trim()}` } : {})
        };
    },

    // Unified request method with global error handling
    async request(endpoint, method = 'GET', body = null) {
        const config = {
            method,
            headers: this.getHeaders()
        };
        if (body) config.body = JSON.stringify(body);

        try {
            const response = await fetch(`${this.baseUrl}${endpoint}`, config);
            
            // Session expired handler
            if (response.status === 401) {
                logout();
                throw new Error('Session expired. Please log in again.');
            }

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Server error: ${response.status}`);
            }

            // Return null for "No Content", otherwise parse JSON
            return response.status === 204 ? null : response.json();
        } catch (err) {
            console.error(`[API ${method}] ${endpoint} failed:`, err.message);
            throw err;
        }
    }
};

// --- 2. AUTHENTICATION LOGIC ---

let currentAuthMode = 'login';

/**
 * Handle Login and Signup form submission
 */
async function handleAuth(e) {
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
        const data = await API.request(endpoint, 'POST', payload);

        if (currentAuthMode === 'login') {
            const token = data.token || data.access_token;
            if (token) {
                localStorage.setItem('token', token);
                showSection('todo');
            }
        } else {
            alert('Account created successfully! Please sign in.');
            toggleAuthMode();
        }
    } catch (err) {
        alert(err.message);
    }
}

function logout() {
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- 3. TASK MANAGEMENT ---

/**
 * Fetch and refresh the task list
 */
async function loadTasks() {
    try {
        const tasks = await API.request('/tasks');
        renderTasks(tasks);
    } catch (err) {
        // Errors are handled by API wrapper
    }
}

/**
 * Send a request to create a new task
 */
async function createTask() {
    const titleInput = document.getElementById('task-title');
    const title = titleInput.value.trim();
    if (!title) return;

    try {
        await API.request('/tasks/create', 'POST', { title });
        titleInput.value = '';
        await loadTasks();
    } catch (err) {
        alert(err.message);
    }
}

/**
 * Update task status (start/done)
 */
async function updateTaskStatus(taskId, action) {
    try {
        await API.request(`/tasks/${action}`, 'POST', { id: taskId });
        await loadTasks();
    } catch (err) {
        console.error('Update status failed');
    }
}

/**
 * Delete a task permanently
 */
async function deleteTask(taskId) {
    if (!confirm("Are you sure you want to delete this task?")) return;
    try {
        await API.request('/tasks/delete', 'DELETE', { id: taskId });
        await loadTasks();
    } catch (err) {
        alert('Delete failed');
    }
}

// --- 4. UI RENDERING & NAVIGATION ---

/**
 * Switch visibility between main sections
 */
function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    sections.forEach(s => document.getElementById(s)?.classList.add('d-none'));

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
        if (!localStorage.getItem('token')) return showSection('welcome');
        document.getElementById('todo-section')?.classList.remove('d-none');
        document.getElementById('nav-user')?.classList.remove('d-none');
        loadTasks();
    }
}

/**
 * Render task items into the list container
 */
function renderTasks(tasks) {
    const list = document.getElementById('tasks-list');
    if (!list) return;

    list.innerHTML = tasks.length === 0 
        ? '<div class="text-center p-3 text-muted">No tasks yet. Enjoy your day!</div>' 
        : '';

    tasks.forEach(task => {
        const item = document.createElement('div');
        item.className = 'list-group-item d-flex justify-content-between align-items-center shadow-sm mb-3 border-0 rounded p-3';
        
        const isDone = task.status === 'done';
        const isInProgress = task.status === 'in_progress';
        const badgeClass = isDone ? 'bg-success' : (isInProgress ? 'bg-warning text-dark' : 'bg-primary');

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
}

/**
 * Switch Auth UI labels (Login vs Signup)
 */
function updateAuthUI() {
    const signupFields = document.getElementById('signup-fields');
    const title = document.getElementById('auth-title');
    const submitBtn = document.getElementById('auth-submit-btn');
    const switchLink = document.getElementById('auth-switch-link');

    if (currentAuthMode === 'signup') {
        signupFields?.classList.remove('d-none');
        title.innerText = 'Create Your Account';
        submitBtn.innerText = 'Register';
        switchLink.innerText = 'Already have an account? Sign In';
    } else {
        signupFields?.classList.add('d-none');
        title.innerText = 'Welcome Back';
        submitBtn.innerText = 'Sign In';
        switchLink.innerText = 'New here? Create an Account';
    }
}

function toggleAuthMode() {
    currentAuthMode = (currentAuthMode === 'login') ? 'signup' : 'login';
    updateAuthUI();
}

// --- 5. INITIALIZATION ---

window.onload = () => {
    // Event listeners
    document.getElementById('auth-form')?.addEventListener('submit', handleAuth);
    
    // Check initial state
    const token = localStorage.getItem('token');
    token ? showSection('todo') : showSection('welcome');
};

// Expose functions globally for HTML onclick events
window.updateTaskStatus = updateTaskStatus;
window.deleteTask = deleteTask;
window.showSection = showSection;
window.toggleAuthMode = toggleAuthMode;
window.logout = logout;
window.createTask = createTask;