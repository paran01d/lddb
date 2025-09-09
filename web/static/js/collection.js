// Collection-specific functionality

// Enhanced collection management
class CollectionManager {
    constructor() {
        this.currentPage = 0;
        this.itemsPerPage = 20;
        this.sortBy = 'title';
        this.sortOrder = 'asc';
        this.filterWatched = 'all'; // all, watched, unwatched
    }

    // Load collection with advanced filtering and sorting
    async loadCollection(options = {}) {
        const {
            search = currentSearch,
            offset = 0,
            limit = this.itemsPerPage,
            sortBy = this.sortBy,
            sortOrder = this.sortOrder,
            filterWatched = this.filterWatched
        } = options;

        try {
            showLoading();
            
            const params = new URLSearchParams({
                limit: limit,
                offset: offset
            });

            if (search) {
                params.append('search', search);
            }

            const data = await apiCall(`/collection?${params}`);
            let laserdiscs = data.laserdiscs || [];
            
            // Client-side filtering for watched status
            if (filterWatched === 'watched') {
                laserdiscs = laserdiscs.filter(ld => ld.watched);
            } else if (filterWatched === 'unwatched') {
                laserdiscs = laserdiscs.filter(ld => !ld.watched);
            }

            // Client-side sorting
            laserdiscs.sort((a, b) => {
                let aVal = a[sortBy] || '';
                let bVal = b[sortBy] || '';
                
                if (typeof aVal === 'string') {
                    aVal = aVal.toLowerCase();
                    bVal = bVal.toLowerCase();
                }

                if (sortOrder === 'desc') {
                    return bVal < aVal ? -1 : bVal > aVal ? 1 : 0;
                } else {
                    return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
                }
            });

            collection = laserdiscs;
            currentOffset = offset;

            updateStats(data.stats);
            this.renderCollection();
            this.updatePagination(data.pagination);
            
        } catch (error) {
            elements.collection.innerHTML = '<p class="error">Failed to load collection. Please try again.</p>';
        }
    }

    // Enhanced collection rendering with advanced features
    renderCollection() {
        if (collection.length === 0) {
            elements.collection.innerHTML = `
                <div class="empty-state">
                    <h3>📀 No LaserDiscs found</h3>
                    <p>${currentSearch ? 'Try a different search term' : 'Start by scanning a barcode or adding a LaserDisc manually'}</p>
                    <button class="primary-btn" onclick="openModal('scan')">Scan Your First LaserDisc</button>
                </div>
            `;
            return;
        }

        // Create collection controls
        const controls = this.createCollectionControls();
        
        // Create grid
        const grid = document.createElement('div');
        grid.className = 'laserdisc-grid';

        collection.forEach(laserdisc => {
            const card = this.createEnhancedLaserDiscCard(laserdisc);
            grid.appendChild(card);
        });

        elements.collection.innerHTML = '';
        elements.collection.appendChild(controls);
        elements.collection.appendChild(grid);
    }

    // Create collection controls (sort, filter, view options)
    createCollectionControls() {
        const controls = document.createElement('div');
        controls.className = 'collection-controls';
        controls.innerHTML = `
            <div class="controls-row">
                <div class="filter-group">
                    <label>Filter:</label>
                    <select id="filter-watched" onchange="collectionManager.updateFilter(this.value)">
                        <option value="all">All LaserDiscs</option>
                        <option value="unwatched">Unwatched</option>
                        <option value="watched">Watched</option>
                    </select>
                </div>
                
                <div class="sort-group">
                    <label>Sort by:</label>
                    <select id="sort-by" onchange="collectionManager.updateSort(this.value)">
                        <option value="title">Title</option>
                        <option value="year">Year</option>
                        <option value="director">Director</option>
                        <option value="added_date">Date Added</option>
                    </select>
                    <button class="sort-order-btn" onclick="collectionManager.toggleSortOrder()" title="Toggle sort order">
                        ${this.sortOrder === 'asc' ? '▲' : '▼'}
                    </button>
                </div>

                <div class="view-group">
                    <button class="view-btn active" onclick="collectionManager.setView('grid')" title="Grid view">⊞</button>
                    <button class="view-btn" onclick="collectionManager.setView('list')" title="List view">☰</button>
                </div>
            </div>
        `;

        // Set current values
        setTimeout(() => {
            const filterSelect = controls.querySelector('#filter-watched');
            const sortSelect = controls.querySelector('#sort-by');
            if (filterSelect) filterSelect.value = this.filterWatched;
            if (sortSelect) sortSelect.value = this.sortBy;
        }, 0);

        return controls;
    }

    // Create enhanced LaserDisc card with more features
    createEnhancedLaserDiscCard(laserdisc) {
        const card = document.createElement('div');
        card.className = 'laserdisc-card';
        card.dataset.id = laserdisc.id;

        const watchedClass = laserdisc.watched ? 'true' : 'false';
        const watchedIcon = laserdisc.watched ? '✅' : '📺';
        const watchedText = laserdisc.watched ? 'Watched' : 'Unwatched';
        const watchBtnText = laserdisc.watched ? 'Mark Unwatched' : 'Mark Watched';

        // Format date
        const addedDate = new Date(laserdisc.added_date).toLocaleDateString();

        card.innerHTML = `
            <div class="card-header">
                ${laserdisc.cover_image_url && laserdisc.cover_image_url !== 'https://www.lddb.com/images/visual/loading.gif' ? 
                    `<div class="card-cover">
                        <img src="${laserdisc.cover_image_url}" alt="${escapeHtml(laserdisc.title)} cover" loading="lazy" 
                             onerror="this.style.display='none'; this.parentNode.classList.add('no-image')">
                    </div>` : ''}
                <div class="card-header-content">
                    <h3>${escapeHtml(laserdisc.title)}</h3>
                    <span class="watched-status ${watchedClass}" title="${watchedText}">
                        ${watchedIcon}
                    </span>
                </div>
            </div>
            
            <div class="card-body">
                <div class="card-info">
                    <p><strong>Year:</strong> ${laserdisc.year || 'Unknown'}</p>
                    <p><strong>UPC:</strong> <code>${laserdisc.upc}</code></p>
                    ${laserdisc.director ? `<p><strong>Director:</strong> ${escapeHtml(laserdisc.director)}</p>` : ''}
                    ${laserdisc.genre ? `<p><strong>Category:</strong> ${escapeHtml(laserdisc.genre)}</p>` : ''}
                    ${laserdisc.format ? `<p><strong>Format:</strong> <span class="format-badge">${laserdisc.format}</span></p>` : ''}
                    ${laserdisc.runtime ? `<p><strong>Runtime:</strong> ${laserdisc.runtime} min</p>` : ''}
                    ${laserdisc.sides ? `<p><strong>Sides:</strong> ${laserdisc.sides}</p>` : ''}
                </div>
                
                ${laserdisc.notes ? `<div class="card-notes"><strong>Notes:</strong> ${escapeHtml(laserdisc.notes)}</div>` : ''}
                
                <div class="card-meta">
                    <small>Added: ${addedDate}</small>
                    ${laserdisc.lddb_url ? `<a href="${laserdisc.lddb_url}" target="_blank" class="lddb-link" title="View on LDDB">🔗 LDDB Details</a>` : ''}
                </div>
            </div>
            
            <div class="card-actions">
                <button class="watch-btn ${laserdisc.watched ? 'watched' : 'unwatched'}" 
                        onclick="toggleWatched(${laserdisc.id})" 
                        title="${watchBtnText}">
                    ${laserdisc.watched ? '👁️' : '👁️‍🗨️'} ${watchBtnText}
                </button>
                <button class="edit-btn" onclick="editLaserDisc(${laserdisc.id})" title="Edit LaserDisc">✏️</button>
                <button class="delete-btn" onclick="deleteLaserDisc(${laserdisc.id})" title="Delete LaserDisc">🗑️</button>
            </div>
        `;

        return card;
    }

    // Update filter
    updateFilter(filter) {
        this.filterWatched = filter;
        this.loadCollection();
    }

    // Update sort
    updateSort(sortBy) {
        this.sortBy = sortBy;
        this.loadCollection();
    }

    // Toggle sort order
    toggleSortOrder() {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
        this.loadCollection();
    }

    // Set view type
    setView(viewType) {
        const grid = document.querySelector('.laserdisc-grid');
        if (grid) {
            grid.className = viewType === 'list' ? 'laserdisc-list' : 'laserdisc-grid';
        }

        // Update active view button
        document.querySelectorAll('.view-btn').forEach(btn => btn.classList.remove('active'));
        event.target.classList.add('active');
    }

    // Update pagination
    updatePagination(pagination) {
        // TODO: Add pagination controls if needed
        console.log('Pagination:', pagination);
    }
}

// Initialize collection manager
const collectionManager = new CollectionManager();

// Replace the original loadCollection function
window.loadCollection = function(search = '', offset = 0) {
    currentSearch = search;
    collectionManager.loadCollection({ search, offset });
};

// Bulk operations
class BulkOperations {
    constructor() {
        this.selectedItems = new Set();
    }

    // Toggle item selection
    toggleSelection(id) {
        if (this.selectedItems.has(id)) {
            this.selectedItems.delete(id);
        } else {
            this.selectedItems.add(id);
        }
        this.updateSelectionUI();
    }

    // Select all visible items
    selectAll() {
        collection.forEach(item => {
            this.selectedItems.add(item.id);
        });
        this.updateSelectionUI();
    }

    // Clear selection
    clearSelection() {
        this.selectedItems.clear();
        this.updateSelectionUI();
    }

    // Update selection UI
    updateSelectionUI() {
        const count = this.selectedItems.size;
        // TODO: Update bulk operation buttons
        console.log(`${count} items selected`);
    }

    // Bulk mark as watched
    async markAllWatched() {
        const promises = Array.from(this.selectedItems).map(id => 
            apiCall(`/collection/${id}/watched`, { method: 'POST' })
        );
        
        try {
            await Promise.all(promises);
            showNotification(`${this.selectedItems.size} items marked as watched`, 'success');
            this.clearSelection();
            loadCollection(currentSearch, currentOffset);
        } catch (error) {
            showNotification('Error updating some items', 'error');
        }
    }

    // Bulk delete
    async deleteSelected() {
        if (!confirm(`Delete ${this.selectedItems.size} selected items?`)) {
            return;
        }

        const promises = Array.from(this.selectedItems).map(id => 
            apiCall(`/collection/${id}`, { method: 'DELETE' })
        );
        
        try {
            await Promise.all(promises);
            showNotification(`${this.selectedItems.size} items deleted`, 'success');
            this.clearSelection();
            loadCollection(currentSearch, currentOffset);
        } catch (error) {
            showNotification('Error deleting some items', 'error');
        }
    }
}

// Initialize bulk operations
const bulkOps = new BulkOperations();

// Export to global scope
window.collectionManager = collectionManager;
window.bulkOps = bulkOps;