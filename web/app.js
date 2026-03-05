/**
 * Configuration and Global State
 */
const API_URL = window.location.origin + '/api';

// DOM Element Selectors
const authSection = document.getElementById('authSection');
const taskSection = document.getElementById('taskSection');
const taskList = document.getElementById('taskList');
const loginBtn = document.getElementById('loginBtn');
const logoutBtn = document.getElementById('logoutBtn');
const addBtn = document.getElementById('addBtn');

/**
 * checkAuth: Toggles visibility between login form and task list 
 * based on the presence of a JWT in localStorage.
 */
function checkAuth() {
    const token = localStorage.getItem('token');
    if (token) {
        authSection.classList.add('d-none');
        taskSection.classList.remove('d-none');
        loadTasks();
    } else {
        authSection.classList.remove('d-none');
        taskSection.classList.add('d-none');
    }
}

/**
 * handleLogin: Authenticates the user and stores the token.
 */
loginBtn.addEventListener('click', async () => {
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });

        if (response.ok) {
            const data = await response.json();
            localStorage.setItem('token', data.token); // Store JWT
            checkAuth();
        } else {
            alert('Authentication failed! Check your credentials.');
        }
    } catch (error) {
        console.error('Login error:', error);
    }
});

/**
 * loadTasks: Fetches the user's task list from the protected API endpoint.
 */
async function loadTasks() {
    const token = localStorage.getItem('token');
    try {
        const response = await fetch(`${API_URL}/tasks`, {
            headers: { 'Authorization': token }
        });

        if (response.ok) {
            const tasks = await response.json();
            renderTasks(tasks);
        } else if (response.status === 401) {
            // Token might be expired
            logout();
        }
    } catch (error) {
        console.error('Failed to fetch tasks:', error);
    }
}

/**
 * renderTasks: Generates HTML for the task list.
 */
function renderTasks(tasks) {
    taskList.innerHTML = '';
    
    // Ensure we have an array
    if (!tasks || tasks.length === 0) {
        taskList.innerHTML = '<li class="list-group-item text-center">No tasks found.</li>';
        return;
    }

    tasks.forEach(task => {
        const li = document.createElement('li');
        li.className = 'list-group-item d-flex justify-content-between align-items-center animate-fade-in';
        
        const isDone = task.status === 'done';
        
        li.innerHTML = `
            <span style="${isDone ? 'text-decoration: line-through; color: gray;' : ''}">
                ${task.title} 
                <span class="badge ${getStatusBadgeClass(task.status)} ms-2">${task.status}</span>
            </span>
            <div>
                ${task.status === 'todo' ? `<button onclick="updateTaskStatus(${task.id}, 'start')" class="btn btn-sm btn-warning">Start</button>` : ''}
                ${task.status === 'in_progress' ? `<button onclick="updateTaskStatus(${task.id}, 'done')" class="btn btn-sm btn-success">Done</button>` : ''}
                <button onclick="deleteTask(${task.id})" class="btn btn-sm btn-outline-danger ms-1">&times;</button>
            </div>
        `;
        taskList.appendChild(li);
    });
}

// Utility to pick badge color
function getStatusBadgeClass(status) {
    switch(status) {
        case 'todo': return 'bg-secondary';
        case 'in_progress': return 'bg-primary';
        case 'done': return 'bg-success';
        default: return 'bg-dark';
    }
}

/**
 * createTask: Sends a POST request to add a new task.
 */
addBtn.addEventListener('click', async () => {
    const titleInput = document.getElementById('taskTitle');
    const token = localStorage.getItem('token');

    if (!titleInput.value) return;

    await fetch(`${API_URL}/tasks/create`, {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json',
            'Authorization': token 
        },
        body: JSON.stringify({ title: titleInput.value })
    });

    titleInput.value = '';
    loadTasks();
});

/**
 * updateTaskStatus: Handles 'start' and 'done' status updates.
 */
async function updateTaskStatus(id, action) {
    const endpoint = action === 'start' ? '/tasks/start' : '/tasks/done';
    await fetch(`${API_URL}${endpoint}`, {
        method: 'POST',
        headers: { 
            'Authorization': localStorage.getItem('token'), 
            'Content-Type': 'application/json' 
        },
        body: JSON.stringify({ id })
    });
    loadTasks();
}

/**
 * deleteTask: Removes a task after confirmation.
 */
async function deleteTask(id) {
    if(!confirm('Are you sure you want to delete this task?')) return;
    
    await fetch(`${API_URL}/tasks/delete`, {
        method: 'DELETE',
        headers: { 
            'Authorization': localStorage.getItem('token'), 
            'Content-Type': 'application/json' 
        },
        body: JSON.stringify({ id })
    });
    loadTasks();
}

/**
 * logout: Clears session and returns to login screen.
 */
function logout() {
    localStorage.removeItem('token');
    checkAuth();
}

logoutBtn.addEventListener('click', logout);

// Application bootstrap
checkAuth();