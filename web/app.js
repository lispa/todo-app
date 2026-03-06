/**
 * Todo-App Frontend Logic
 * Integrated SPA Navigation, JWT Auth, and Task Management
 * Comments: English
 */

const API_URL = '/api';
let currentAuthMode = 'login'; 

// --- SECTION 1: UI NAVIGATION CONTROL ---

/**
 * Handles switching between Welcome, Auth, and Todo sections
 */
function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    
    // Hide all sections initially
    sections.forEach(s => document.getElementById(s)?.classList.add('d-none'));
    
    // Hide user navigation by default
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
        // Redirect to welcome if no token found
        if (!token) return showSection('welcome');
        
        document.getElementById('todo-section')?.classList.remove('d-none');
        document.getElementById('nav-user')?.classList.remove('d-none');
        loadTasks();
    }
}

/**
 * Updates Auth form labels based on Login/Signup mode
 */
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

// --- SECTION 2: AUTHENTICATION ---

/**
 * Handles Form Submission for Login and Registration
 */
document.getElementById('auth-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const endpoint = currentAuthMode === 'signup' ? '/auth/signup' : '/auth/login';
    
    const payload = { email, password };
    if (currentAuthMode === 'signup') {
        payload.first_name = document.getElementById('first_name').value || "User";
        payload.last_name = document.getElementById('last_name').value || "New";
    }

    try {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const data = await response.json();
        console.log("Auth Debug Response:", data);

        if (response.ok) {
            if (currentAuthMode === 'login') {
                // Backend might return 'token' or 'access_token'
                const token = data.token || data.access_token;
                if (token) {
                    localStorage.setItem('token', token);
                    showSection('todo');
                } else {
                    alert("Authentication successful, but no token received.");
                }
            } else {
                alert('Account created! Switching to Login.');
                currentAuthMode = 'login';
                updateAuthUI();
            }
        } else {
            alert(data.error || 'Authentication failed. Please check your credentials.');
        }
    } catch (err) {
        console.error('Network or Server Error:', err);
    }
});

/**
 * Clears local session and returns to welcome screen
 */
function logout() {
    console.log("User logged out. Clearing local storage.");
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- SECTION 3: TASK MANAGEMENT ---

/**
 * Fetches tasks from the API using Bearer Token
 */
async function loadTasks() {
    const token = localStorage.getItem('token');
    if (!token) return logout();

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            method: 'GET',
            headers: { 
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });

        // If server returns 401, the token is likely expired or invalid
        if (response.status === 401) {
            console.error("Session expired (401). Redirecting to login.");
            return logout();
        }

        const tasks = await response.json();
        const list = document.getElementById('tasks-list');
        if (!list) return;

        list.innerHTML = '';
        if (!tasks || tasks.length === 0) {
            list.innerHTML = '<div class="text-center p-3 text-muted">Your task list is empty.</div>';
            return;
        }

        // Render tasks dynamically
        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center shadow-sm mb-2 border-0 rounded animate-fade-in';
            item.innerHTML = `
                <div>
                    <h6 class="mb-0">${task.title}</h6>
                    <small class="text-muted">Status: ${task.status}</small>
                </div>
                <span class="badge bg-primary rounded-pill">${task.status}</span>
            `;
            list.appendChild(item);
        });
    } catch (err) {
        console.error('Failed to fetch tasks:', err);
    }
}

/**
 * Sends a POST request to create a new task
 */
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
            titleInput.value = ''; // Clear input on success
            loadTasks(); // Refresh list
        } else {
            const err = await response.json();
            alert(err.error || "Failed to create task");
        }
    } catch (err) { 
        console.error("Create task request failed:", err); 
    }
}

// --- SECTION 4: APP INITIALIZATION ---

window.onload = () => {
    const token = localStorage.getItem('token');
    console.log("App startup. Session found:", !!token);
    
    // Persistent login check
    token ? showSection('todo') : showSection('welcome');
};