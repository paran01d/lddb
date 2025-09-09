// API base URL - will be set dynamically
const API_BASE = window.location.origin + '/api';

// Global state
let collection = [];
let currentSearch = '';
let currentOffset = 0;
const LIMIT = 20;

// DOM elements
const elements = {
    stats: {
        total: document.getElementById('total-count'),
        watched: document.getElementById('watched-count'),
        unwatched: document.getElementById('unwatched-count')
    },
    buttons: {
        scan: document.getElementById('scan-btn'),
        addManual: document.getElementById('add-manual-btn'),
        random: document.getElementById('random-btn'),
        search: document.getElementById('search-btn')
    },
    inputs: {
        search: document.getElementById('search-input'),
        manualUpc: document.getElementById('manual-upc'),
        lookupBtn: document.getElementById('lookup-btn')
    },
    modals: {
        scan: document.getElementById('scan-modal'),
        add: document.getElementById('add-modal')
    },
    collection: document.getElementById('collection'),
    addForm: document.getElementById('add-form')
};

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    console.log('LDDB Collection Manager starting...');
    initializeEventListeners();
    loadCollection();
    
    // Add debug console for mobile (development only)
    addMobileDebugConsole();
});

// Add mobile debug console
function addMobileDebugConsole() {
    // Only add on mobile or when URL contains debug=1
    const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    const debugMode = window.location.search.includes('debug=1');
    
    if (isMobile || debugMode) {
        const debugConsole = document.createElement('div');
        debugConsole.id = 'mobile-console';
        debugConsole.innerHTML = `
            <div class="console-header">
                <span>Debug Console</span>
                <button onclick="toggleConsole()">−</button>
            </div>
            <div class="console-content" id="console-content"></div>
        `;
        document.body.appendChild(debugConsole);
        
        // Intercept console.log
        const originalLog = console.log;
        console.log = function(...args) {
            originalLog.apply(console, args);
            addToMobileConsole('LOG', args.join(' '));
        };
        
        const originalError = console.error;
        console.error = function(...args) {
            originalError.apply(console, args);
            addToMobileConsole('ERROR', args.join(' '));
        };
    }
}

function addToMobileConsole(type, message) {
    const content = document.getElementById('console-content');
    if (!content) return;
    
    const timestamp = new Date().toLocaleTimeString();
    const logEntry = document.createElement('div');
    logEntry.className = `console-entry console-${type.toLowerCase()}`;
    logEntry.innerHTML = `<span class="timestamp">[${timestamp}]</span> ${message}`;
    
    content.appendChild(logEntry);
    content.scrollTop = content.scrollHeight;
    
    // Keep only last 50 entries
    const entries = content.querySelectorAll('.console-entry');
    if (entries.length > 50) {
        entries[0].remove();
    }
}

function toggleConsole() {
    const console = document.getElementById('mobile-console');
    const content = console.querySelector('.console-content');
    const btn = console.querySelector('button');
    
    if (content.style.display === 'none') {
        content.style.display = 'block';
        btn.textContent = '−';
    } else {
        content.style.display = 'none';
        btn.textContent = '+';
    }
}

// Event listeners
function initializeEventListeners() {
    // Search functionality
    elements.buttons.search.addEventListener('click', handleSearch);
    elements.inputs.search.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            handleSearch();
        }
    });

    // Modal controls
    elements.buttons.scan.addEventListener('click', () => openModal('scan'));
    elements.buttons.addManual.addEventListener('click', () => openModal('add'));
    elements.buttons.random.addEventListener('click', getRandomMovie);

    // UPC lookup
    elements.inputs.lookupBtn.addEventListener('click', lookupUPC);
    elements.inputs.manualUpc.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            lookupUPC();
        }
    });

    // Add form submission
    elements.addForm.addEventListener('submit', handleAddLaserDisc);

    // Close modals when clicking outside or on close button
    document.querySelectorAll('.close').forEach(closeBtn => {
        closeBtn.addEventListener('click', closeModals);
    });

    window.addEventListener('click', function(e) {
        if (e.target.classList.contains('modal')) {
            closeModals();
        }
    });

    // Escape key to close modals
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeModals();
        }
    });
}

// API functions
async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || `HTTP ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        showNotification(`Error: ${error.message}`, 'error');
        throw error;
    }
}

// Load collection from API
async function loadCollection(search = '', offset = 0) {
    try {
        showLoading();
        const params = new URLSearchParams({
            limit: LIMIT,
            offset: offset
        });

        if (search) {
            params.append('search', search);
        }

        const data = await apiCall(`/collection?${params}`);
        
        collection = data.laserdiscs || [];
        currentSearch = search;
        currentOffset = offset;

        updateStats(data.stats);
        renderCollection();
        
    } catch (error) {
        elements.collection.innerHTML = '<p class="error">Failed to load collection. Please try again.</p>';
    }
}

// Render collection in the UI
function renderCollection() {
    if (collection.length === 0) {
        elements.collection.innerHTML = `
            <div class="empty-state">
                <h3>No LaserDiscs found</h3>
                <p>${currentSearch ? 'Try a different search term' : 'Start by scanning a barcode or adding a LaserDisc manually'}</p>
            </div>
        `;
        return;
    }

    const grid = document.createElement('div');
    grid.className = 'laserdisc-grid';

    collection.forEach(laserdisc => {
        const card = createLaserDiscCard(laserdisc);
        grid.appendChild(card);
    });

    elements.collection.innerHTML = '';
    elements.collection.appendChild(grid);
}

// Create a LaserDisc card element
function createLaserDiscCard(laserdisc) {
    const card = document.createElement('div');
    card.className = 'laserdisc-card';
    card.dataset.id = laserdisc.id;

    const watchedClass = laserdisc.watched ? 'true' : 'false';
    const watchedText = laserdisc.watched ? '✅ Watched' : '📺 Unwatched';
    const watchBtnText = laserdisc.watched ? 'Mark Unwatched' : 'Mark Watched';

    card.innerHTML = `
        <h3>${escapeHtml(laserdisc.title)}</h3>
        <p><strong>Year:</strong> ${laserdisc.year || 'Unknown'}</p>
        <p><strong>UPC:</strong> ${laserdisc.upc}</p>
        ${laserdisc.director ? `<p><strong>Director:</strong> ${escapeHtml(laserdisc.director)}</p>` : ''}
        ${laserdisc.genre ? `<p><strong>Genre:</strong> ${escapeHtml(laserdisc.genre)}</p>` : ''}
        ${laserdisc.format ? `<p><strong>Format:</strong> ${laserdisc.format}</p>` : ''}
        ${laserdisc.runtime ? `<p><strong>Runtime:</strong> ${laserdisc.runtime} min</p>` : ''}
        ${laserdisc.notes ? `<p><strong>Notes:</strong> ${escapeHtml(laserdisc.notes)}</p>` : ''}
        <span class="watched ${watchedClass}">${watchedText}</span>
        
        <div class="card-actions">
            <button class="watch-btn" onclick="toggleWatched(${laserdisc.id})">${watchBtnText}</button>
            <button class="edit-btn" onclick="editLaserDisc(${laserdisc.id})">Edit</button>
            <button class="delete-btn" onclick="deleteLaserDisc(${laserdisc.id})">Delete</button>
        </div>
    `;

    return card;
}

// Update statistics display
function updateStats(stats) {
    elements.stats.total.textContent = stats.total || 0;
    elements.stats.watched.textContent = stats.watched || 0;
    elements.stats.unwatched.textContent = stats.unwatched || 0;
}

// Search functionality
function handleSearch() {
    const searchTerm = elements.inputs.search.value.trim();
    loadCollection(searchTerm, 0);
}

// UPC lookup functionality
async function lookupUPC() {
    const upc = elements.inputs.manualUpc.value.trim();
    if (!upc) {
        showNotification('Please enter a UPC', 'error');
        return;
    }

    try {
        showNotification('Looking up LaserDisc...', 'info');
        const data = await apiCall(`/lookup/${encodeURIComponent(upc)}`);

        if (data.result && data.result.found) {
            // Pre-fill the add form with the lookup data
            populateAddForm(data.result);
            closeModals();
            openModal('add');
            showNotification('LaserDisc found! Review and add to collection.', 'success');
        } else {
            showNotification('LaserDisc not found in database', 'warning');
        }
    } catch (error) {
        // Error already handled in apiCall
    }
}

// Populate add form with lookup data
function populateAddForm(data) {
    document.getElementById('form-upc').value = data.upc || '';
    document.getElementById('form-title').value = data.title || '';
    document.getElementById('form-year').value = data.year || '';
    document.getElementById('form-director').value = data.director || '';
    document.getElementById('form-genre').value = data.genre || '';
    document.getElementById('form-format').value = data.format || '';
    document.getElementById('form-sides').value = data.sides || '';
    document.getElementById('form-runtime').value = data.runtime || '';
}

// Handle add LaserDisc form submission
async function handleAddLaserDisc(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const laserdisc = {
        upc: formData.get('upc') || document.getElementById('form-upc').value,
        title: formData.get('title') || document.getElementById('form-title').value,
        year: parseInt(formData.get('year') || document.getElementById('form-year').value) || 0,
        director: formData.get('director') || document.getElementById('form-director').value || '',
        genre: formData.get('genre') || document.getElementById('form-genre').value || '',
        format: formData.get('format') || document.getElementById('form-format').value || '',
        sides: parseInt(formData.get('sides') || document.getElementById('form-sides').value) || 0,
        runtime: parseInt(formData.get('runtime') || document.getElementById('form-runtime').value) || 0,
        notes: formData.get('notes') || document.getElementById('form-notes').value || ''
    };

    if (!laserdisc.upc || !laserdisc.title) {
        showNotification('UPC and Title are required', 'error');
        return;
    }

    try {
        await apiCall('/collection', {
            method: 'POST',
            body: JSON.stringify(laserdisc)
        });

        showNotification('LaserDisc added successfully!', 'success');
        closeModals();
        resetAddForm();
        loadCollection(currentSearch, 0); // Refresh collection
    } catch (error) {
        // Error already handled in apiCall
    }
}

// Toggle watched status
async function toggleWatched(id) {
    try {
        await apiCall(`/collection/${id}/watched`, {
            method: 'POST'
        });

        showNotification('Watched status updated', 'success');
        loadCollection(currentSearch, currentOffset); // Refresh current view
    } catch (error) {
        // Error already handled in apiCall
    }
}

// Delete LaserDisc
async function deleteLaserDisc(id) {
    if (!confirm('Are you sure you want to delete this LaserDisc?')) {
        return;
    }

    try {
        await apiCall(`/collection/${id}`, {
            method: 'DELETE'
        });

        showNotification('LaserDisc deleted successfully', 'success');
        loadCollection(currentSearch, currentOffset); // Refresh current view
    } catch (error) {
        // Error already handled in apiCall
    }
}

// Get random unwatched movie
async function getRandomMovie() {
    try {
        const data = await apiCall('/random-unwatched');
        
        if (data.laserdisc) {
            showRandomMovieModal(data.laserdisc);
        }
    } catch (error) {
        if (error.message.includes('No unwatched')) {
            showNotification('No unwatched LaserDiscs found!', 'info');
        }
    }
}

// Show random movie modal
function showRandomMovieModal(laserdisc) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <span class="close">&times;</span>
            <h2>🎲 Random Movie Selected!</h2>
            <div class="random-movie">
                <h3>${escapeHtml(laserdisc.title)}</h3>
                <p><strong>Year:</strong> ${laserdisc.year || 'Unknown'}</p>
                ${laserdisc.director ? `<p><strong>Director:</strong> ${escapeHtml(laserdisc.director)}</p>` : ''}
                ${laserdisc.genre ? `<p><strong>Genre:</strong> ${escapeHtml(laserdisc.genre)}</p>` : ''}
                ${laserdisc.runtime ? `<p><strong>Runtime:</strong> ${laserdisc.runtime} min</p>` : ''}
                <div class="random-actions">
                    <button class="primary-btn" onclick="markAsWatched(${laserdisc.id})">Mark as Watched</button>
                    <button class="secondary-btn" onclick="closeModals()">Pick Another</button>
                </div>
            </div>
        </div>
    `;

    document.body.appendChild(modal);
    modal.style.display = 'block';

    // Close button
    modal.querySelector('.close').addEventListener('click', () => {
        document.body.removeChild(modal);
    });
}

// Mark as watched from random modal
async function markAsWatched(id) {
    await toggleWatched(id);
    closeModals();
}

// Modal functions
function openModal(modalType) {
    closeModals();
    elements.modals[modalType].style.display = 'block';
}

function closeModals() {
    Object.values(elements.modals).forEach(modal => {
        modal.style.display = 'none';
    });
    
    // Remove any random movie modals
    document.querySelectorAll('.modal').forEach(modal => {
        if (modal.querySelector('.random-movie')) {
            modal.remove();
        }
    });
}

// Reset add form
function resetAddForm() {
    elements.addForm.reset();
}

// Utility functions
function showLoading() {
    elements.collection.innerHTML = '<p class="loading">Loading collection...</p>';
}

function showNotification(message, type = 'info') {
    // Remove existing notification
    const existing = document.querySelector('.notification');
    if (existing) {
        existing.remove();
    }

    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // Auto remove after 3 seconds
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 3000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Global functions for inline event handlers
window.toggleWatched = toggleWatched;
window.deleteLaserDisc = deleteLaserDisc;
window.editLaserDisc = function(id) {
    // TODO: Implement edit functionality
    showNotification('Edit functionality coming soon!', 'info');
};
window.markAsWatched = markAsWatched;
window.closeModals = closeModals;