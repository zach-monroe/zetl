// ============================================
// Zetl - Main JavaScript
// ============================================

// Get current user ID from data attribute (set by Go template)
const currentUserId = document.body.dataset.userId ? parseInt(document.body.dataset.userId) : null;

// Track currently hovered card for flip-during-hover detection
let currentlyHoveredCard = null;

// ============================================
// Card Flip Functionality
// ============================================
function initCardFlip() {
  document.querySelectorAll('.card-footer').forEach(footer => {
    footer.addEventListener('click', (e) => {
      e.stopPropagation();
      const card = footer.closest('.quote-card');
      // Add has-flipped class to enable animations after first interaction
      if (!card.classList.contains('has-flipped')) {
        card.classList.add('has-flipped');
      }
      card.classList.toggle('flipped');

      // If card is currently hovered, recalculate expansion for new face
      if (currentlyHoveredCard === card) {
        // Small delay to let the flip class toggle take effect
        setTimeout(() => {
          expandCard(card);
        }, 50);
      }
    });
  });
}

// ============================================
// Card Menu Functionality
// ============================================
function initCardMenus() {
  // Prevent card flip when clicking on menu
  document.querySelectorAll('.card-menu-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      // Close other menus
      document.querySelectorAll('.card-menu').forEach(menu => {
        if (menu !== btn.nextElementSibling) {
          menu.classList.remove('open');
        }
      });
      btn.nextElementSibling.classList.toggle('open');
    });
  });

  // Close menus when clicking outside
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.card-menu-container')) {
      document.querySelectorAll('.card-menu').forEach(menu => {
        menu.classList.remove('open');
      });
    }
  });
}

// ============================================
// Smooth Card Expansion Animation
// ============================================
const CARD_BASE_HEIGHT = 280;
const ANIMATION_DURATION = 500; // ms
const BOUNCE_EASING = 'cubic-bezier(0.34, 1.56, 0.64, 1)'; // Bouncy for expansion
const SMOOTH_EASING = 'cubic-bezier(0.4, 0, 0.2, 1)'; // Smooth ease-out for contraction

// Store original positions for FLIP animation
let cardPositions = new Map();

// Get all card positions
function captureCardPositions() {
  const positions = new Map();
  document.querySelectorAll('.quote-card').forEach(card => {
    const rect = card.getBoundingClientRect();
    positions.set(card, { top: rect.top, left: rect.left });
  });
  return positions;
}

// Animate cards from old positions to new positions (FLIP technique)
function animateCardsToNewPositions(oldPositions) {
  document.querySelectorAll('.quote-card').forEach(card => {
    const oldPos = oldPositions.get(card);
    if (!oldPos) return;

    const newRect = card.getBoundingClientRect();
    const deltaY = oldPos.top - newRect.top;
    const deltaX = oldPos.left - newRect.left;

    // Skip if no movement
    if (Math.abs(deltaY) < 1 && Math.abs(deltaX) < 1) return;

    // Mark as animating to disable hover scale
    card.classList.add('animating');

    // Apply inverse transform to start from old position
    card.style.transform = `translate(${deltaX}px, ${deltaY}px) scale(1)`;
    card.style.transition = 'none';

    // Force reflow
    card.offsetHeight;

    // Animate to new position
    card.style.transition = `transform ${ANIMATION_DURATION}ms ${BOUNCE_EASING}`;
    card.style.transform = '';

    // Remove animating class after animation completes
    setTimeout(() => {
      card.classList.remove('animating');
      card.style.transition = '';
    }, ANIMATION_DURATION);
  });
}

// Measure expanded height of a card
function measureExpandedHeight(card) {
  const cardInner = card.querySelector('.card-inner');
  const cardFront = card.querySelector('.card-front');
  const cardBack = card.querySelector('.card-back');
  const isFlipped = card.classList.contains('flipped');
  const activeFace = isFlipped ? cardBack : cardFront;

  // Store original styles
  const originalInnerHeight = cardInner.style.height;
  const originalFrontHeight = cardFront.style.height;
  const originalBackHeight = cardBack.style.height;
  const originalBackPosition = cardBack.style.position;

  // Temporarily remove height constraints and make back face relative for measurement
  cardInner.style.height = 'auto';
  cardFront.style.height = 'auto';
  cardBack.style.height = 'auto';

  // Temporarily make the active face position relative to get accurate measurement
  if (isFlipped) {
    cardBack.style.position = 'relative';
  }

  // Measure the natural height of the active face
  const expandedHeight = Math.max(activeFace.scrollHeight, CARD_BASE_HEIGHT);

  // Restore original styles
  cardInner.style.height = originalInnerHeight;
  cardFront.style.height = originalFrontHeight;
  cardBack.style.height = originalBackHeight;
  if (isFlipped) {
    cardBack.style.position = originalBackPosition;
  }

  return expandedHeight;
}

// Expand a card to fit its content (can be called on hover or flip)
function expandCard(card) {
  const cardInner = card.querySelector('.card-inner');
  const cardFront = card.querySelector('.card-front');

  // Capture positions of all cards before expansion
  const oldPositions = captureCardPositions();

  // Measure what the expanded height should be for the current face
  const expandedHeight = measureExpandedHeight(card);

  // Get current height for smooth transition
  const currentHeight = cardInner.offsetHeight || CARD_BASE_HEIGHT;

  // Choose easing based on whether we're expanding or contracting
  const isExpanding = expandedHeight > currentHeight;
  const easing = isExpanding ? BOUNCE_EASING : SMOOTH_EASING;

  // Set starting heights explicitly
  cardInner.style.height = `${currentHeight}px`;
  cardFront.style.height = `${currentHeight}px`;

  // Force reflow
  card.offsetHeight;

  // Set up transitions
  cardInner.style.transition = `height ${ANIMATION_DURATION}ms ${easing}`;
  cardFront.style.transition = `height ${ANIMATION_DURATION}ms ${easing}`;

  // Animate to expanded height
  cardInner.style.height = `${expandedHeight}px`;
  cardFront.style.height = `${expandedHeight}px`;

  // Animate other cards to their new positions
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      animateCardsToNewPositions(oldPositions);
    });
  });
}

// Handle card expansion on hover
function initCardExpansion() {
  document.querySelectorAll('.quote-card').forEach(card => {
    const cardInner = card.querySelector('.card-inner');
    const cardFront = card.querySelector('.card-front');

    let expandTimeout = null;

    card.addEventListener('mouseenter', () => {
      if (expandTimeout) clearTimeout(expandTimeout);

      // Track this card as hovered
      currentlyHoveredCard = card;

      // Expand the card
      expandCard(card);
    });

    card.addEventListener('mouseleave', () => {
      // Clear hover tracking
      currentlyHoveredCard = null;

      expandTimeout = setTimeout(() => {
        // Capture positions before collapse
        const oldPositions = captureCardPositions();

        // Use smooth easing for collapse (no bounce)
        cardInner.style.transition = `height ${ANIMATION_DURATION}ms ${SMOOTH_EASING}`;
        cardFront.style.transition = `height ${ANIMATION_DURATION}ms ${SMOOTH_EASING}`;

        // Animate back to base height
        cardInner.style.height = `${CARD_BASE_HEIGHT}px`;
        cardFront.style.height = `${CARD_BASE_HEIGHT}px`;

        // Animate other cards
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            animateCardsToNewPositions(oldPositions);
          });
        });

        // Clean up explicit heights after animation
        setTimeout(() => {
          if (currentlyHoveredCard !== card) {
            cardInner.style.height = '';
            cardFront.style.height = '';
            cardInner.style.transition = '';
            cardFront.style.transition = '';
          }
        }, ANIMATION_DURATION + 50);
      }, 50); // Small delay to prevent flickering on quick mouse movements
    });
  });
}

// ============================================
// Session Management
// ============================================
async function checkSession() {
  try {
    const response = await fetch('/api/user', {
      method: 'GET',
      credentials: 'same-origin'
    });
    return response.ok;
  } catch {
    return false;
  }
}

// ============================================
// Edit Modal Functions
// ============================================
function openEditModal(quoteId, quote, author, book, tags, notes) {
  document.getElementById('edit-quote-id').value = quoteId;
  document.getElementById('edit-quote-text').value = quote;
  document.getElementById('edit-author').value = author;
  document.getElementById('edit-book').value = book || '';
  document.getElementById('edit-tags').value = tags || '';
  document.getElementById('edit-notes').value = notes || '';
  document.getElementById('edit-error').classList.add('hidden');

  const modal = document.getElementById('edit-modal');
  modal.classList.add('open');
  document.body.style.overflow = 'hidden';
}

function closeEditModal() {
  const modal = document.getElementById('edit-modal');
  modal.classList.remove('open');
  document.body.style.overflow = '';
}

// ============================================
// Delete Modal Functions
// ============================================
function openDeleteModal(quoteId) {
  document.getElementById('delete-quote-id').value = quoteId;
  document.getElementById('delete-error').classList.add('hidden');

  const modal = document.getElementById('delete-modal');
  modal.classList.add('open');
  document.body.style.overflow = 'hidden';
}

function closeDeleteModal() {
  const modal = document.getElementById('delete-modal');
  modal.classList.remove('open');
  document.body.style.overflow = '';
}

// Delete Confirmation
async function confirmDelete() {
  const errorDiv = document.getElementById('delete-error');
  errorDiv.classList.add('hidden');

  // Check session first
  if (!await checkSession()) {
    errorDiv.textContent = 'Your session has expired. Please log in again.';
    errorDiv.classList.remove('hidden');
    return;
  }

  const quoteId = document.getElementById('delete-quote-id').value;

  try {
    const response = await fetch(`/api/quote/${quoteId}`, {
      method: 'DELETE',
      credentials: 'same-origin'
    });

    if (response.ok) {
      closeDeleteModal();
      // Remove the card from DOM
      const card = document.querySelector(`[data-quote-id="${quoteId}"]`);
      if (card) {
        card.style.transform = 'scale(0.8)';
        card.style.opacity = '0';
        setTimeout(() => card.remove(), 300);
      }
    } else {
      const data = await response.json();
      errorDiv.textContent = data.error || 'Failed to delete quote.';
      errorDiv.classList.remove('hidden');
    }
  } catch (error) {
    errorDiv.textContent = 'An error occurred. Please try again.';
    errorDiv.classList.remove('hidden');
  }
}

// ============================================
// Add Quote Modal Functions
// ============================================
function openAddModal() {
  // Clear the form
  document.getElementById('add-form').reset();
  document.getElementById('add-error').classList.add('hidden');

  const modal = document.getElementById('add-modal');
  modal.classList.add('open');
  document.body.style.overflow = 'hidden';

  // Focus on the quote textarea
  setTimeout(() => {
    document.getElementById('add-quote-text').focus();
  }, 100);
}

function closeAddModal() {
  const modal = document.getElementById('add-modal');
  modal.classList.remove('open');
  document.body.style.overflow = '';
}

// ============================================
// Form Submission Handlers
// ============================================
function initFormHandlers() {
  // Edit Form Submit
  document.getElementById('edit-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorDiv = document.getElementById('edit-error');
    errorDiv.classList.add('hidden');

    // Check session first
    if (!await checkSession()) {
      errorDiv.textContent = 'Your session has expired. Please log in again.';
      errorDiv.classList.remove('hidden');
      return;
    }

    const quoteId = document.getElementById('edit-quote-id').value;
    const tagsStr = document.getElementById('edit-tags').value;
    const tags = tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(t => t) : [];

    const formData = {
      quote: document.getElementById('edit-quote-text').value,
      author: document.getElementById('edit-author').value,
      book: document.getElementById('edit-book').value,
      tags: tags,
      notes: document.getElementById('edit-notes').value
    };

    try {
      const response = await fetch(`/api/quote/${quoteId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify(formData)
      });

      const data = await response.json();

      if (response.ok) {
        closeEditModal();
        // Refresh the page to show updated content
        window.location.reload();
      } else {
        errorDiv.textContent = data.error || 'Failed to update quote.';
        errorDiv.classList.remove('hidden');
      }
    } catch (error) {
      errorDiv.textContent = 'An error occurred. Please try again.';
      errorDiv.classList.remove('hidden');
    }
  });

  // Add Form Submit
  document.getElementById('add-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorDiv = document.getElementById('add-error');
    errorDiv.classList.add('hidden');

    // Check session first
    if (!await checkSession()) {
      errorDiv.textContent = 'Your session has expired. Please log in again.';
      errorDiv.classList.remove('hidden');
      return;
    }

    const tagsStr = document.getElementById('add-tags').value;
    const tags = tagsStr ? tagsStr.split(',').map(t => t.trim()).filter(t => t) : [];

    const formData = {
      quote: document.getElementById('add-quote-text').value,
      author: document.getElementById('add-author').value,
      book: document.getElementById('add-book').value,
      tags: tags,
      notes: document.getElementById('add-notes').value
    };

    try {
      const response = await fetch('/api/quote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify(formData)
      });

      const data = await response.json();

      if (response.ok) {
        closeAddModal();
        // Refresh the page to show the new quote
        window.location.reload();
      } else {
        errorDiv.textContent = data.error || 'Failed to add quote.';
        errorDiv.classList.remove('hidden');
      }
    } catch (error) {
      errorDiv.textContent = 'An error occurred. Please try again.';
      errorDiv.classList.remove('hidden');
    }
  });
}

// ============================================
// Modal Event Handlers
// ============================================
function initModalHandlers() {
  // Close modals on overlay click
  document.querySelectorAll('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) {
        overlay.classList.remove('open');
        document.body.style.overflow = '';
      }
    });
  });

  // Close modals on Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeEditModal();
      closeDeleteModal();
      closeAddModal();
    }
  });
}

// ============================================
// Search and Tag Filtering
// ============================================
let activeTagFilters = new Set();
let searchQuery = '';

// Store card data for filtering
const cardData = [];
const allTags = new Set();

function initCardData() {
  document.querySelectorAll('.quote-card').forEach(card => {
    const quoteText = card.querySelector('.quote-text')?.textContent?.trim() || '';
    const author = card.querySelector('.text-cyan-400')?.textContent?.trim() || '';
    const book = card.querySelector('.text-zinc-500.italic')?.textContent?.trim() || '';
    const tags = Array.from(card.querySelectorAll('.tag-pill')).map(t => t.textContent.trim().toLowerCase());

    // Collect unique tags
    tags.forEach(tag => allTags.add(tag));

    cardData.push({
      element: card,
      quote: quoteText.toLowerCase(),
      author: author.toLowerCase(),
      book: book.toLowerCase(),
      tags: tags
    });
  });
}

// Fuzzy match function - returns score (higher = better match), -1 for no match
function fuzzyMatch(pattern, str) {
  pattern = pattern.toLowerCase();
  str = str.toLowerCase();

  // Exact match gets highest score
  if (str === pattern) return 1000;

  // Starts with pattern gets high score
  if (str.startsWith(pattern)) return 500 + (pattern.length / str.length) * 100;

  // Contains pattern gets medium score
  if (str.includes(pattern)) return 200 + (pattern.length / str.length) * 100;

  // Fuzzy character matching
  let patternIdx = 0;
  let score = 0;
  let consecutiveBonus = 0;

  for (let i = 0; i < str.length && patternIdx < pattern.length; i++) {
    if (str[i] === pattern[patternIdx]) {
      score += 10 + consecutiveBonus;
      consecutiveBonus += 5; // Reward consecutive matches
      patternIdx++;
    } else {
      consecutiveBonus = 0;
    }
  }

  // All characters must match
  if (patternIdx === pattern.length) {
    return score;
  }

  return -1; // No match
}

// Helper function to escape HTML
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Render tag search results
function renderTagSearchResults(query = '') {
  const tagList = document.getElementById('filter-tag-list');
  if (!tagList) return;

  if (allTags.size === 0) {
    tagList.innerHTML = '<p class="text-zinc-500 text-sm italic px-3 py-2">No tags available</p>';
    return;
  }

  let tagsToShow;

  if (query.trim() === '') {
    // Show all tags sorted alphabetically when no query
    tagsToShow = Array.from(allTags)
      .filter(tag => !activeTagFilters.has(tag.toLowerCase().trim()))
      .sort()
      .slice(0, 10); // Limit to 10 when showing all
  } else {
    // Fuzzy search and sort by score
    tagsToShow = Array.from(allTags)
      .filter(tag => !activeTagFilters.has(tag.toLowerCase().trim()))
      .map(tag => ({ tag, score: fuzzyMatch(query, tag) }))
      .filter(item => item.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 8)
      .map(item => item.tag);
  }

  if (tagsToShow.length === 0) {
    tagList.innerHTML = query.trim()
      ? '<p class="text-zinc-500 text-sm italic px-3 py-2">No matching tags</p>'
      : '<p class="text-zinc-500 text-sm italic px-3 py-2">All tags selected</p>';
    return;
  }

  tagList.innerHTML = tagsToShow.map(tag => `
    <div class="filter-tag-result" data-tag="${escapeHtml(tag)}" onclick="selectTagFromSearch('${escapeHtml(tag)}')">
      <svg class="w-4 h-4 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A2 2 0 013 12V7a4 4 0 014-4z"/>
      </svg>
      <span class="text-zinc-300 text-sm">${highlightMatch(tag, query)}</span>
    </div>
  `).join('');
}

// Highlight matching characters in tag
function highlightMatch(tag, query) {
  if (!query.trim()) return escapeHtml(tag);

  const lowerTag = tag.toLowerCase();
  const lowerQuery = query.toLowerCase();

  // Simple highlight for contains match
  const idx = lowerTag.indexOf(lowerQuery);
  if (idx !== -1) {
    return escapeHtml(tag.slice(0, idx)) +
           '<span class="text-cyan-400 font-medium">' + escapeHtml(tag.slice(idx, idx + query.length)) + '</span>' +
           escapeHtml(tag.slice(idx + query.length));
  }

  return escapeHtml(tag);
}

// Select tag from search results
function selectTagFromSearch(tag) {
  const normalizedTag = tag.toLowerCase().trim();
  if (!activeTagFilters.has(normalizedTag)) {
    activeTagFilters.add(normalizedTag);
    renderFilterChips();
    updateSelectedTagsSection();
    updateFilterBadge();
    applyFilters();

    // Clear search and refresh results
    const searchInput = document.getElementById('tag-search-input');
    if (searchInput) {
      searchInput.value = '';
      renderTagSearchResults('');
    }
  }
}

// Update selected tags section in dropdown
function updateSelectedTagsSection() {
  const section = document.getElementById('selected-tags-section');
  const list = document.getElementById('selected-tags-list');
  if (!section || !list) return;

  if (activeTagFilters.size === 0) {
    section.classList.add('hidden');
    return;
  }

  section.classList.remove('hidden');
  list.innerHTML = Array.from(activeTagFilters).map(tag => `
    <span class="selected-tag-chip" onclick="removeTagFromDropdown('${escapeHtml(tag)}')">
      ${escapeHtml(tag)}
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
      </svg>
    </span>
  `).join('');
}

// Remove tag from dropdown
function removeTagFromDropdown(tag) {
  const normalizedTag = tag.toLowerCase().trim();
  activeTagFilters.delete(normalizedTag);
  renderFilterChips();
  updateSelectedTagsSection();
  updateFilterBadge();
  applyFilters();

  // Refresh search results to show the removed tag again
  const searchInput = document.getElementById('tag-search-input');
  renderTagSearchResults(searchInput?.value || '');
}

// Legacy function for backward compatibility
function toggleTagFromDropdown(tag) {
  selectTagFromSearch(tag);
}

// Legacy function
function updateFilterDropdown() {
  const searchInput = document.getElementById('tag-search-input');
  renderTagSearchResults(searchInput?.value || '');
  updateSelectedTagsSection();
}

// Renamed from populateFilterDropdown
function populateFilterDropdown() {
  renderTagSearchResults('');
  updateSelectedTagsSection();
}

// Update filter badge count
function updateFilterBadge() {
  const badge = document.getElementById('filter-badge');
  if (!badge) return;

  const count = activeTagFilters.size;
  if (count > 0) {
    badge.textContent = count;
    badge.classList.remove('hidden');
    badge.classList.add('flex');
  } else {
    badge.classList.add('hidden');
    badge.classList.remove('flex');
  }
}

// Add tag filter
function addTagFilter(tag) {
  const normalizedTag = tag.toLowerCase().trim();
  if (!activeTagFilters.has(normalizedTag)) {
    activeTagFilters.add(normalizedTag);
    renderFilterChips();
    updateFilterDropdown();
    updateFilterBadge();
    applyFilters();
  }
}

// Remove tag filter
function removeTagFilter(tag) {
  activeTagFilters.delete(tag);
  renderFilterChips();
  updateFilterDropdown();
  updateFilterBadge();
  applyFilters();
}

// Clear all filters
function clearAllFilters() {
  activeTagFilters.clear();
  searchQuery = '';
  const searchInputEl = document.getElementById('search-input');
  if (searchInputEl) searchInputEl.value = '';
  renderFilterChips();
  updateFilterDropdown();
  updateFilterBadge();
  applyFilters();
}

// Render filter chips
function renderFilterChips() {
  const container = document.getElementById('filter-chips');
  const filtersContainer = document.getElementById('active-filters');

  if (!container || !filtersContainer) return;

  container.innerHTML = '';

  if (activeTagFilters.size === 0) {
    filtersContainer.classList.add('hidden');
    return;
  }

  filtersContainer.classList.remove('hidden');

  activeTagFilters.forEach(tag => {
    const chip = document.createElement('span');
    chip.className = 'inline-flex items-center gap-1 px-3 py-1 bg-cyan-900/50 border border-cyan-700 text-cyan-300 rounded-full text-sm';
    chip.innerHTML = `
      ${escapeHtml(tag)}
      <button type="button" onclick="removeTagFilter('${escapeHtml(tag)}')" class="ml-1 hover:text-white transition-colors">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
        </svg>
      </button>
    `;
    container.appendChild(chip);
  });
}

// Apply all filters
function applyFilters() {
  let visibleCount = 0;

  cardData.forEach(data => {
    let matchesSearch = true;
    let matchesTags = true;

    // Check search query
    if (searchQuery) {
      matchesSearch = data.quote.includes(searchQuery) ||
                     data.author.includes(searchQuery) ||
                     data.book.includes(searchQuery);
    }

    // Check tag filters (AND logic - card must have ALL selected tags)
    if (activeTagFilters.size > 0) {
      matchesTags = Array.from(activeTagFilters).every(tag =>
        data.tags.includes(tag)
      );
    }

    if (matchesSearch && matchesTags) {
      data.element.style.display = '';
      visibleCount++;
    } else {
      data.element.style.display = 'none';
    }
  });

  // Show/hide no results message
  const noResults = document.getElementById('no-results');
  const quotesGrid = document.getElementById('quotes-grid');

  if (visibleCount === 0 && (searchQuery || activeTagFilters.size > 0)) {
    noResults?.classList.remove('hidden');
    quotesGrid?.classList.add('hidden');
  } else {
    noResults?.classList.add('hidden');
    quotesGrid?.classList.remove('hidden');
  }
}

// Initialize search functionality
function initSearch() {
  const searchInputEl = document.getElementById('search-input');
  if (searchInputEl) {
    searchInputEl.addEventListener('input', (e) => {
      searchQuery = e.target.value.toLowerCase().trim();
      applyFilters();
    });
  }

  // Initialize tag search input
  const tagSearchInput = document.getElementById('tag-search-input');
  if (tagSearchInput) {
    tagSearchInput.addEventListener('input', (e) => {
      renderTagSearchResults(e.target.value);
    });

    // Prevent dropdown from closing when clicking in search input
    tagSearchInput.addEventListener('click', (e) => {
      e.stopPropagation();
    });
  }

  // Initialize filter dropdown
  populateFilterDropdown();
}

// Hide quote function (for non-owner cards)
function hideQuote(quoteId) {
  const card = document.querySelector(`[data-quote-id="${quoteId}"]`);
  if (card) {
    // Animate out
    card.style.transition = 'all 0.3s ease';
    card.style.transform = 'scale(0.8)';
    card.style.opacity = '0';
    setTimeout(() => {
      card.style.display = 'none';
      // Also remove from cardData so it doesn't appear in filter results
      const index = cardData.findIndex(c => c.element === card);
      if (index > -1) {
        cardData.splice(index, 1);
      }
    }, 300);
  }
}

// ============================================
// Global Exports (for onclick handlers in HTML)
// ============================================
window.addTagFilter = addTagFilter;
window.removeTagFilter = removeTagFilter;
window.clearAllFilters = clearAllFilters;
window.toggleTagFromDropdown = toggleTagFromDropdown;
window.selectTagFromSearch = selectTagFromSearch;
window.removeTagFromDropdown = removeTagFromDropdown;
window.openAddModal = openAddModal;
window.closeAddModal = closeAddModal;
window.openEditModal = openEditModal;
window.closeEditModal = closeEditModal;
window.openDeleteModal = openDeleteModal;
window.closeDeleteModal = closeDeleteModal;
window.confirmDelete = confirmDelete;
window.hideQuote = hideQuote;

// ============================================
// Initialize Everything
// ============================================
document.addEventListener('DOMContentLoaded', () => {
  initCardFlip();
  initCardMenus();
  initCardExpansion();
  initFormHandlers();
  initModalHandlers();
  initCardData();
  initSearch();
});
