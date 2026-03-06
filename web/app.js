/**
 * Todo-App Frontend Logic
 * Handles SPA navigation, JWT authentication, and task management.
 * @version 1.2
 */

const API_URL = '/api';
let currentAuthMode = 'login'; 

// --- SECTION 1: NAVIGATION ---

/**
 * Switch between application sections: Welcome, Auth, or Dashboard.
 */
function showSection(section, mode = null) {
    const sections = ['welcome-section', 'auth-section', 'todo-section'];
    
    // Hide all main containers
    sections.forEach(s => {
        const el = document.getElementById(s);
        if (el) el.classList.add('d-none');
    });

    // Reset user-specific navigation
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
        // If no token exists, force redirect to landing page
        if (!token) {
            console.warn("No session token found. Redirecting to welcome.");
            return showSection('welcome');
        }
        
        document.getElementById('todo-section')?.classList.remove('d-none');
        document.getElementById('nav-user')?.classList.remove('d-none');
        loadTasks(); // Load data from backend
    }
}

/**
 * Update UI labels and fields based on login or registration mode.
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

/**
 * Handle form submission for both login and signup.
 */
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
        console.log(`Sending ${currentAuthMode} request to ${endpoint}...`);
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const data = await response.json();
        console.log("Auth Debug - Server Response Data:", data);

        if (response.ok) {
            if (currentAuthMode === 'login') {
                // IMPORTANT: Backend must return field 'token' or 'access_token'
                const token = data.token || data.access_token;
                if (token) {
                    localStorage.setItem('token', token);
                    console.log("Session saved. Entering dashboard.");
                    showSection('todo');
                } else {
                    console.error("Auth successful but token missing in response body!");
                    alert("System error: Token not received.");
                }
            } else {
                alert('Account created! Please sign in with your new credentials.');
                currentAuthMode = 'login';
                updateAuthUI();
            }
        } else {
            console.error("Auth failed:", data.error);
            alert(data.error || 'Authentication failed. Please check your inputs.');
        }
    } catch (err) {
        console.error('Critical Auth Error:', err);
    }
});

/**
 * Remove session data and return to welcome screen.
 */
function logout() {
    console.log("User logged out. Clearing local storage.");
    localStorage.removeItem('token');
    showSection('welcome');
}

// --- SECTION 3: TASK MANAGEMENT ---

/**
 * Fetch tasks using the JWT Bearer token.
 */
async function loadTasks() {
    const token = localStorage.getItem('token');
    if (!token) return showSection('welcome');

    console.log("Loading tasks with token:", token.substring(0, 10) + "...");

    try {
        const response = await fetch(`${API_URL}/tasks`, {
            method: 'GET',
            headers: { 
                'Authorization': `Bearer ${token.trim()}`,
                'Accept': 'application/json'
            }
        });

        if (response.status === 401) {
            console.error("Server returned 401 Unauthorized. Potential JWT_SECRET mismatch on backend.");
            // Temporary: don't logout to allow debugging
            // return logout(); 
            return;
        }

        const tasks = await response.json();
        const list = document.getElementById('tasks-list');
        if (!list) return;

        list.innerHTML = tasks.length === 0 ? '<div class="text-center p-3 text-muted">No tasks yet.</div>' : '';

        tasks.forEach(task => {
            const item = document.createElement('div');
            item.className = 'list-group-item d-flex justify-content-between align-items-center shadow-sm mb-2 border-0 rounded';
            item.innerHTML = `
                <div>
                    <h6 class="mb-0 text-dark">${task.title}</h6>
                    <small class="text-muted">Status: ${task.status}</small>
                </div>
                <span class="badge bg-primary rounded-pill px-3">${task.status}</span>
            `;
            list.appendChild(item);
        });
    } catch (err) {
        console.error('Task fetch failed:', err);
    }
}

/**
 * Create a new task entry in the database.
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
            titleInput.value = '';
            loadTasks(); // Reload the list
        } else {
            const error = await response.json();
            console.error("Task creation failed:", error);
        }
    } catch (err) { 
        console.error("Network error during task creation:", err); 
    }
}

// --- SECTION 4: INITIALIZATION ---

/**
 * On page load, check for existing session.
 */
window.onload = () => {
    const token = localStorage.getItem('token');
    console.log("App bootstrap. Session found:", !!token);
    
    token ? showSection('todo') : showSection('welcome');
};