// Public collection display (read-only)

let collection = [];
let currentSearch = '';
let currentOffset = 0;

// Simple API call for public endpoint (no auth required)
async function apiCall(endpoint) {
    try {
        const response = await fetch(`/api/public${endpoint}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// HTML escape function
function escapeHtml(unsafe) {
    if (!unsafe) return '';
    return unsafe
        .toString()
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// Public collection manager (read-only)
class PublicCollectionManager {
    constructor() {
        this.currentPage = 0;
        this.itemsPerPage = 20;
        this.sortBy = 'title';
        this.sortOrder = 'asc';
        this.filterWatched = 'all'; // all, watched, unwatched
        this.isLoading = false;
        this.hasMoreData = true;
        this.allLoadedData = [];
        this.currentView = 'grid'; // grid or list
        this.setupInfiniteScroll();
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Search
        const searchInput = document.getElementById('search-input');
        const searchBtn = document.getElementById('search-btn');

        if (searchInput) {
            searchInput.addEventListener('keyup', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch();
                }
            });
        }

        if (searchBtn) {
            searchBtn.addEventListener('click', () => this.performSearch());
        }

        // Filter
        const filterSelect = document.getElementById('filter-select');
        if (filterSelect) {
            filterSelect.addEventListener('change', (e) => {
                this.filterWatched = e.target.value;
                this.loadCollection({ offset: 0, append: false });
            });
        }

        // Sort
        const sortSelect = document.getElementById('sort-select');
        if (sortSelect) {
            sortSelect.addEventListener('change', (e) => {
                this.sortBy = e.target.value;
                this.loadCollection({ offset: 0, append: false });
            });
        }

        // View toggle
        const viewToggle = document.getElementById('view-toggle');
        if (viewToggle) {
            viewToggle.addEventListener('click', () => {
                this.currentView = this.currentView === 'grid' ? 'list' : 'grid';
                viewToggle.textContent = this.currentView === 'grid' ? 'Grid' : 'List';

                // Update the grid class
                const grid = document.getElementById('laserdisc-grid');
                if (grid) {
                    grid.className = this.currentView === 'list' ? 'laserdisc-list' : 'laserdisc-grid';
                }
            });
        }
    }

    performSearch() {
        const searchInput = document.getElementById('search-input');
        currentSearch = searchInput ? searchInput.value.trim() : '';
        this.loadCollection({ offset: 0, append: false });
    }

    async loadCollection(options = {}) {
        const {
            search = currentSearch,
            offset = 0,
            limit = this.itemsPerPage,
            sortBy = this.sortBy,
            sortOrder = this.sortOrder,
            filterWatched = this.filterWatched,
            append = false
        } = options;

        if (this.isLoading) return;
        this.isLoading = true;

        try {
            if (!append) {
                this.showLoading();
                this.allLoadedData = [];
                this.hasMoreData = true;
                this.currentPage = 0;
            }

            const params = new URLSearchParams({
                limit: limit,
                offset: offset
            });

            if (search) {
                params.append('search', search);
            }

            const data = await apiCall(`/collection?${params}`);
            const newLaserdiscs = data.laserdiscs || [];

            // Check if we have more data
            this.hasMoreData = newLaserdiscs.length === limit && data.pagination.total > (offset + limit);

            // Store all loaded data
            if (append) {
                this.allLoadedData = [...this.allLoadedData, ...newLaserdiscs];
            } else {
                this.allLoadedData = newLaserdiscs;
            }

            // Apply client-side filtering and sorting
            let filteredData = [...this.allLoadedData];

            // Client-side filtering for watched status
            if (filterWatched === 'watched') {
                filteredData = filteredData.filter(ld => ld.watched);
            } else if (filterWatched === 'unwatched') {
                filteredData = filteredData.filter(ld => !ld.watched);
            }

            // Client-side sorting
            filteredData.sort((a, b) => {
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

            collection = filteredData;
            currentOffset = offset;

            this.updateStats(data.stats);

            // When appending, filter and render only the new items
            if (append) {
                // Apply the same filter to new items
                let newFilteredItems = newLaserdiscs;
                if (filterWatched === 'watched') {
                    newFilteredItems = newLaserdiscs.filter(ld => ld.watched);
                } else if (filterWatched === 'unwatched') {
                    newFilteredItems = newLaserdiscs.filter(ld => !ld.watched);
                }
                this.renderCollection(true, newFilteredItems);
            } else {
                this.renderCollection(false);
            }

            this.updatePagination(data.pagination);

            if (!append) {
                this.ensureLoadMoreButton();
            } else {
                this.updateLoadMoreButton();
            }

        } catch (error) {
            console.error('Failed to load collection:', error);
            this.showError('Failed to load collection. Please try again.');
        } finally {
            this.isLoading = false;
        }
    }

    showLoading() {
        const collectionDiv = document.getElementById('collection');
        if (collectionDiv) {
            collectionDiv.innerHTML = '<p class="loading">Loading collection...</p>';
        }
    }

    showError(message) {
        const collectionDiv = document.getElementById('collection');
        if (collectionDiv) {
            collectionDiv.innerHTML = `<p class="error">${escapeHtml(message)}</p>`;
        }
    }

    updateStats(stats) {
        if (!stats) return;

        const totalCount = document.getElementById('total-count');
        const watchedCount = document.getElementById('watched-count');
        const unwatchedCount = document.getElementById('unwatched-count');

        if (totalCount) totalCount.textContent = stats.total || 0;
        if (watchedCount) watchedCount.textContent = stats.watched || 0;
        if (unwatchedCount) unwatchedCount.textContent = stats.unwatched || 0;
    }

    renderCollection(append = false, itemsToRender = null) {
        const collectionDiv = document.getElementById('collection');
        if (!collectionDiv) return;

        if (collection.length === 0 && !append) {
            collectionDiv.innerHTML = '<p class="no-results">No laserdiscs found.</p>';
            return;
        }

        const items = itemsToRender || collection;

        if (!append) {
            // Create grid/list container
            const gridClass = this.currentView === 'list' ? 'laserdisc-list' : 'laserdisc-grid';
            const grid = document.createElement('div');
            grid.className = gridClass;
            grid.id = 'laserdisc-grid';

            items.forEach(laserdisc => {
                const card = this.createLaserDiscCard(laserdisc);
                grid.appendChild(card);
            });

            collectionDiv.innerHTML = '';
            collectionDiv.appendChild(grid);
        } else {
            // Append to existing grid
            const grid = document.getElementById('laserdisc-grid');
            if (grid) {
                // Remove the load more button before appending new items
                const loadMoreBtn = collectionDiv.querySelector('.load-more-container');
                if (loadMoreBtn) {
                    loadMoreBtn.remove();
                }

                items.forEach(laserdisc => {
                    const card = this.createLaserDiscCard(laserdisc);
                    grid.appendChild(card);
                });
            }
        }
    }

    createLaserDiscCard(laserdisc) {
        const card = document.createElement('div');
        card.className = 'laserdisc-card public-card';
        card.dataset.id = laserdisc.id;

        const watchedClass = laserdisc.watched ? 'true' : 'false';
        const watchedIcon = laserdisc.watched ? '✅' : '📺';
        const watchedText = laserdisc.watched ? 'Watched' : 'Unwatched';

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
        `;

        return card;
    }

    setupInfiniteScroll() {
        const throttle = (func, delay) => {
            let timeout;
            return (...args) => {
                clearTimeout(timeout);
                timeout = setTimeout(() => func.apply(this, args), delay);
            };
        };

        const handleScroll = throttle(() => {
            if (this.isLoading || !this.hasMoreData) return;

            const scrollPosition = window.scrollY + window.innerHeight;
            const documentHeight = document.documentElement.scrollHeight;
            const threshold = 200;

            if (scrollPosition >= documentHeight - threshold) {
                this.loadMore();
            }
        }, 300);

        window.addEventListener('scroll', handleScroll);
    }

    loadMore() {
        if (this.isLoading || !this.hasMoreData) return;

        this.currentPage++;
        const newOffset = this.currentPage * this.itemsPerPage;

        this.loadCollection({
            offset: newOffset,
            append: true
        });
    }

    ensureLoadMoreButton() {
        const collectionDiv = document.getElementById('collection');
        if (!collectionDiv) return;

        let loadMoreContainer = collectionDiv.querySelector('.load-more-container');

        if (!loadMoreContainer) {
            loadMoreContainer = document.createElement('div');
            loadMoreContainer.className = 'load-more-container';
            loadMoreContainer.innerHTML = `
                <button class="load-more-btn" onclick="collectionManager.loadMore()">
                    Load More
                </button>
            `;
            collectionDiv.appendChild(loadMoreContainer);
        }

        this.updateLoadMoreButton();
    }

    updateLoadMoreButton() {
        const loadMoreContainer = document.getElementById('collection')?.querySelector('.load-more-container');
        if (!loadMoreContainer) return;

        if (this.hasMoreData) {
            loadMoreContainer.style.display = 'block';
            const btn = loadMoreContainer.querySelector('.load-more-btn');
            if (btn) {
                btn.disabled = this.isLoading;
                btn.textContent = this.isLoading ? 'Loading...' : 'Load More';
            }
        } else {
            loadMoreContainer.style.display = 'none';
        }
    }

    updatePagination(pagination) {
        // Update load more button based on pagination info
        if (pagination) {
            // Calculate if there's more data based on pagination info
            const currentlyLoaded = pagination.offset + pagination.limit;
            this.hasMoreData = currentlyLoaded < pagination.total;
            this.updateLoadMoreButton();
        }
    }
}

// Initialize when DOM is ready
let collectionManager;

document.addEventListener('DOMContentLoaded', () => {
    collectionManager = new PublicCollectionManager();
    collectionManager.loadCollection();
});
