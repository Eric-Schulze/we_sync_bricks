/**
 * Minifig Parts Management - Component-based approach
 * This file handles all minifig parts functionality using individual template components
 */

// Global state management
window.partOutValues = {};
window.currentMinifigData = null;

/**
 * Toggle parts list visibility
 */
function togglePartsList(itemId) {
    const partsList = document.getElementById('parts-list-' + itemId);
    const toggleIcon = document.getElementById('toggle-icon-' + itemId);
    const toggleButton = document.getElementById('toggle-parts-' + itemId);
    
    if (partsList.classList.contains('hidden')) {
        partsList.classList.remove('hidden');
        partsList.classList.add('block');
        toggleIcon.textContent = '▼';
        toggleButton.querySelector('span:nth-child(2)').textContent = 'Hide Parts List';
    } else {
        partsList.classList.remove('block');
        partsList.classList.add('hidden');
        toggleIcon.textContent = '▶';
        toggleButton.querySelector('span:nth-child(2)').textContent = 'Show Parts List';
    }
}

/**
 * Load all minifig details (picture, pricing, parts)
 */
function loadMinifigDetails(itemId) {
    console.log('Loading minifig details for:', itemId);
    loadMinifigPicture(itemId);
    loadMinifigPricing(itemId);
    loadMinifigParts(itemId);
}

/**
 * Initialize BrickLink item (called from template)
 */
function initializeBricklinkItem(itemId) {
    if (itemId) {
        loadMinifigDetails(itemId);
    }
}

/**
 * Load minifig picture
 */
function loadMinifigPicture(itemId) {
    console.log('Loading picture for:', itemId);
    
    const pictureContainer = document.getElementById('picture-' + itemId);
    if (!pictureContainer) {
        console.error('Picture container not found for:', itemId);
        return;
    }
    
    console.log('Picture container found:', pictureContainer);
    
    fetch('/partial-minifigs-lists/minifig-picture', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'bricklink_id=' + encodeURIComponent(itemId)
    })
    .then(response => {
        console.log('Picture API response status:', response.status);
        return response.json();
    })
    .then(data => {
        console.log('Picture API response data:', data);
        
        if (data && data.meta && data.meta.code === 200 && data.data) {
            const firstImage = data.data;
            const imageUrl = firstImage.thumbnail_url || firstImage.url;
            if (imageUrl) {
                console.log('Setting image URL:', imageUrl);
                pictureContainer.innerHTML = '<img src="' + imageUrl + '" alt="Minifig ' + itemId + '" class="max-w-full max-h-full object-contain rounded">';
                pictureContainer.classList.add('flex', 'items-center', 'justify-center');
                pictureContainer.offsetHeight;
            } else {
                console.log('No image URL found in response');
                pictureContainer.innerHTML = '<div class="text-center text-gray-400"><div class="text-2xl">📷</div><div class="text-xs">No URL</div></div>';
            }
        } else {
            console.log('API call failed or no data:', data);
            pictureContainer.innerHTML = '<div class="text-center text-gray-400"><div class="text-2xl">📷</div><div class="text-xs">No image</div></div>';
        }
    })
    .catch(error => {
        console.error('Error loading picture:', error);
        document.getElementById('picture-' + itemId).innerHTML = '<div class="text-center text-red-500"><div class="text-2xl">❌</div><div class="text-xs">Error</div></div>';
    });
}

/**
 * Load minifig pricing (both new and used)
 */
function loadMinifigPricing(itemId) {
    console.log('Loading pricing for:', itemId);
    
    Promise.all([
        loadMinifigPricingByCondition(itemId, 'N'),
        loadMinifigPricingByCondition(itemId, 'U')
    ])
    .then(([newData, usedData]) => {
        updatePriceDisplay(itemId, 'price-new-' + itemId, newData.price);
        updatePriceDisplay(itemId, 'price-used-' + itemId, usedData.price);
        updateSoldCountDisplay(itemId, 'sold-new-' + itemId, newData.total_qty);
        updateSoldCountDisplay(itemId, 'sold-used-' + itemId, usedData.total_qty);
        updatePieceInfoDisplay(itemId, newData.total_qty, usedData.total_qty);
    })
    .catch(error => {
        console.error('Error loading pricing:', error);
        document.getElementById('price-new-' + itemId).textContent = 'Error';
        document.getElementById('price-used-' + itemId).textContent = 'Error';
        updateSoldCountDisplay(itemId, 'sold-new-' + itemId, 0);
        updateSoldCountDisplay(itemId, 'sold-used-' + itemId, 0);
        updatePieceInfoDisplay(itemId, 0, 0);
    });
}

/**
 * Load pricing for specific condition
 */
function loadMinifigPricingByCondition(itemId, condition) {
    return fetch('/partial-minifigs-lists/minifig-pricing', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'bricklink_id=' + encodeURIComponent(itemId) + '&condition=' + encodeURIComponent(condition)
    })
    .then(response => response.json())
    .then(data => {
        if (data && data.meta && data.meta.code === 200 && data.data) {
            let priceItems = Array.isArray(data.data) ? data.data : [data.data];
            if (priceItems.length > 0) {
                const priceItem = priceItems[0];
                const avgPrice = priceItem.qty_avg_price || priceItem.average_price || priceItem.price;
                const totalQty = priceItem.total_quantity || priceItem.quantity_sold || priceItem.qty_sold || 0;
                return {
                    price: avgPrice ? parseFloat(avgPrice) : null,
                    total_qty: totalQty ? parseInt(totalQty) : 0
                };
            }
        }
        return { price: null, total_qty: 0 };
    });
}

/**
 * Update price display element
 */
function updatePriceDisplay(itemId, elementId, price) {
    const element = document.getElementById(elementId);
    if (element) {
        if (price !== null && !isNaN(price)) {
            element.textContent = '$' + price.toFixed(2);
        } else {
            element.textContent = 'N/A';
        }
    }
}

/**
 * Update sold count display element
 */
function updateSoldCountDisplay(itemId, elementId, count) {
    const element = document.getElementById(elementId);
    if (element) {
        if (count && !isNaN(count) && count > 0) {
            element.textContent = '(' + count + ')';
        } else {
            element.textContent = '(0)';
        }
    }
}

/**
 * Update piece information display
 */
function updatePieceInfoDisplay(itemId, newCount, usedCount) {
    const newElement = document.getElementById('piece-count-new-' + itemId);
    const usedElement = document.getElementById('piece-count-used-' + itemId);
    
    if (newElement) {
        if (newCount && !isNaN(newCount) && newCount > 0) {
            newElement.textContent = newCount + ' sold';
        } else {
            newElement.textContent = 'No data';
        }
    }
    
    if (usedElement) {
        if (usedCount && !isNaN(usedCount) && usedCount > 0) {
            usedElement.textContent = usedCount + ' sold';
        } else {
            usedElement.textContent = 'No data';
        }
    }
}

/**
 * Load minifig parts and process data
 */
function loadMinifigParts(itemId) {
    console.log('Loading parts for:', itemId);
    
    fetch('/partial-minifigs-lists/minifig-parts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'bricklink_id=' + encodeURIComponent(itemId)
    })
    .then(response => response.json())
    .then(data => {
        if (data && data.meta && data.meta.code === 200 && data.data && data.data.length > 0) {
            const processedParts = processPartsData(data.data);
            storePartsData(itemId, processedParts);
            initializePartOutTracking(itemId, processedParts);
            enableAddButton(itemId);
        } else {
            clearPartsData(itemId);
            console.log('Parts API error:', data && data.meta ? data.meta.message : 'Unknown error');
        }
    })
    .catch(error => {
        console.error('Error loading parts:', error);
        clearPartsData(itemId);
    });
}

/**
 * Process raw parts data into standardized format
 */
function processPartsData(rawData) {
    const processedParts = [];
    
    rawData.forEach((part, index) => {
        let partData = extractPartData(part);
        if (partData.PartNo) {
            processedParts.push(partData);
        }
    });
    
    return processedParts;
}

/**
 * Extract part data from various API response formats
 */
function extractPartData(part) {
    let partName = 'Unknown Part';
    let partQuantity = 1;
    let partNo = '';
    let colorId = null;
    let itemType = 'PART';
    
    // Check different possible data structures
    if (part.item && part.item.name) {
        partName = part.item.name;
        partNo = part.item.no || '';
        itemType = part.item.type || 'PART';
        partQuantity = part.quantity || 1;
        colorId = part.color_id || null;
    } else if (part.entries && part.entries.length > 0) {
        const entry = part.entries[0];
        if (entry.item && entry.item.name) {
            partName = entry.item.name;
            partNo = entry.item.no || '';
            itemType = entry.item.type || 'PART';
            partQuantity = entry.quantity || 1;
            colorId = entry.color_id || null;
        }
    } else if (part.name) {
        partName = part.name;
        partNo = part.no || '';
        itemType = part.type || 'PART';
        partQuantity = part.quantity || 1;
        colorId = part.color_id || null;
    }
    
    return {
        PartNo: partNo,
        PartName: partName,
        Quantity: partQuantity,
        ColorId: colorId || 0,
        ItemType: itemType
    };
}

/**
 * Store parts data for later use
 */
function storePartsData(itemId, parts) {
    const partsDataField = document.getElementById('parts-data-' + itemId);
    if (partsDataField) {
        partsDataField.value = JSON.stringify(parts);
        console.log('Stored', parts.length, 'parts for', itemId);
    }
}

/**
 * Clear parts data
 */
function clearPartsData(itemId) {
    const partsDataField = document.getElementById('parts-data-' + itemId);
    if (partsDataField) {
        partsDataField.value = '';
    }
}

/**
 * Initialize part out values tracking
 */
function initializePartOutTracking(itemId, parts) {
    window.partOutValues[itemId] = {
        newTotal: 0,
        usedTotal: 0,
        partsLoaded: 0,
        totalParts: parts.length
    };
    
    // Load pricing for part out calculation
    parts.forEach((part, index) => {
        setTimeout(() => {
            loadPartOutPricing(itemId, part.ItemType, part.PartNo, part.ColorId, part.Quantity);
        }, index * 100);
    });
}

/**
 * Load part pricing for part out calculation
 */
function loadPartOutPricing(itemId, itemType, partNo, colorId, quantity) {
    Promise.all([
        loadPartPricingByCondition('temp', itemType, partNo, colorId, 'N'),
        loadPartPricingByCondition('temp', itemType, partNo, colorId, 'U')
    ])
    .then(([newPrice, usedPrice]) => {
        updatePartOutValues(itemId, newPrice, usedPrice, quantity);
    })
    .catch(error => {
        console.error('Error loading part out pricing:', error);
    });
}

/**
 * Load individual part pricing by condition
 */
function loadPartPricingByCondition(partId, itemType, partNo, colorId, condition) {
    return fetch('/partial-minifigs-lists/part-pricing', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'item_type=' + encodeURIComponent(itemType) + 
              '&item_id=' + encodeURIComponent(partNo) + 
              '&color_id=' + encodeURIComponent(colorId) + 
              '&condition=' + encodeURIComponent(condition)
    })
    .then(response => response.json())
    .then(data => {
        if (data && data.meta && data.meta.code === 200 && data.data) {
            let priceItems = Array.isArray(data.data) ? data.data : [data.data];
            if (priceItems.length > 0) {
                const priceItem = priceItems[0];
                const avgPrice = priceItem.avg_price || priceItem.average_price || priceItem.price;
                if (avgPrice) {
                    return parseFloat(avgPrice);
                }
            }
        }
        return null;
    })
    .catch(error => {
        console.error('Error loading part pricing:', error);
        return null;
    });
}

/**
 * Update part out values
 */
function updatePartOutValues(itemId, newPrice, usedPrice, quantity) {
    if (!window.partOutValues[itemId]) {
        window.partOutValues[itemId] = {
            newTotal: 0,
            usedTotal: 0,
            partsLoaded: 0,
            totalParts: 0
        };
    }
    
    const partOutData = window.partOutValues[itemId];
    
    // Add to totals (multiply by quantity)
    if (newPrice !== null && !isNaN(newPrice)) {
        partOutData.newTotal += newPrice * quantity;
    }
    if (usedPrice !== null && !isNaN(usedPrice)) {
        partOutData.usedTotal += usedPrice * quantity;
    }
    
    partOutData.partsLoaded++;
    
    // Update display
    updatePriceDisplay(itemId, 'partout-new-' + itemId, partOutData.newTotal);
    updatePriceDisplay(itemId, 'partout-used-' + itemId, partOutData.usedTotal);
}

/**
 * Enable the add minifig button
 */
function enableAddButton(itemId) {
    const addButton = document.getElementById('add-minifig-' + itemId);
    if (addButton) {
        addButton.disabled = false;
        addButton.textContent = 'Add';
    }
}

/**
 * Add minifig with stored parts data
 */
function addMinifigWithStoredParts(itemId) {
    console.log('Button clicked for minifig:', itemId);
    
    const partsDataField = document.getElementById('parts-data-' + itemId);
    let parts = null;
    
    if (partsDataField && partsDataField.value) {
        try {
            parts = JSON.parse(partsDataField.value);
            console.log('Retrieved stored parts data for', itemId, ':', parts.length, 'parts');
        } catch (error) {
            console.error('Error parsing stored parts data:', error);
            parts = null;
        }
    } else {
        console.log('No parts data found for', itemId);
    }
    
    addMinifigAndParts(itemId, parts);
}

/**
 * Add minifig and parts (main entry point)
 */
function addMinifigAndParts(itemId, parts = null) {
    // Get the list ID from the search results container
    const searchResults = document.getElementById('search-results');
    const listId = searchResults ? searchResults.getAttribute('data-list-id') : null;
    
    if (!listId) {
        alert('Unable to determine which list to add to. Please try again.');
        return;
    }
    
    console.log('Preparing to show minifig details modal for:', itemId);
    
    // Store minifig data for modal use
    window.currentMinifigData = {
        itemId: itemId,
        listId: listId,
        selectedParts: []
    };
    
    showMinifigDetailsModal(parts);
}

/**
 * Show minifig details modal using server-side template
 */
function showMinifigDetailsModal(parts = null) {
    const minifigData = window.currentMinifigData;
    if (!minifigData || !minifigData.itemId) {
        console.error('No minifig data available');
        return;
    }
    
    console.log('Showing modal for:', minifigData.itemId, 'with parts:', parts);
    
    // Use HTMX to load the modal from server
    const modalContainer = document.getElementById('modal-container');
    if (!modalContainer) {
        console.error('Modal container not found');
        // Create modal container if it doesn't exist
        const container = document.createElement('div');
        container.id = 'modal-container';
        document.body.appendChild(container);
    }
    
    // Make HTMX request to load modal
    const url = '/partial-minifigs-lists/minifig-details-modal';
    const formData = new FormData();
    formData.append('item_id', minifigData.itemId);
    formData.append('list_id', minifigData.listId);
    if (parts) {
        formData.append('parts', JSON.stringify(parts));
    }
    
    fetch(url, {
        method: 'POST',
        body: formData
    })
    .then(response => response.text())
    .then(html => {
        const container = document.getElementById('modal-container') || document.body;
        container.innerHTML = html;
        
        // If parts weren't provided, load them via API
        if (!parts) {
            populateModalPartsList();
        }
    })
    .catch(error => {
        console.error('Error loading modal:', error);
        // Fallback: show a simple modal
        showFallbackModal();
    });
}

/**
 * Show fallback modal if server-side loading fails
 */
function showFallbackModal() {
    const minifigData = window.currentMinifigData;
    const modalHtml = `
        <div class="modal-overlay fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" id="minifig-details-overlay" onclick="if(event.target === this) closeMinifigDetailsModal()">
            <div class="modal bg-white rounded-lg p-6 max-w-4xl w-full mx-4 shadow-2xl relative">
                <button type="button" onclick="closeMinifigDetailsModal()" class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors z-10">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                    </svg>
                </button>
                
                <h2 class="m-0 mb-6 text-gray-700 text-center text-xl font-semibold pr-8">Minifig Details</h2>
                
                <div class="text-center py-8">
                    <div class="text-lg font-medium text-gray-700 mb-4">Minifig: ${minifigData.itemId}</div>
                    <div id="modal-parts-list" class="text-gray-500 mb-6">Loading parts...</div>
                    
                    <div class="space-y-4">
                        <div>
                            <label class="block mb-2 font-medium text-gray-700">Reference ID / Location:</label>
                            <input type="text" id="reference_id" name="reference_id" placeholder="e.g., Box A, Shelf 2, etc." class="w-full p-3 border border-gray-300 rounded-md text-sm">
                        </div>
                        <div>
                            <label class="block mb-2 font-medium text-gray-700">Condition:</label>
                            <select id="condition" name="condition" class="w-full p-3 border border-gray-300 rounded-md text-sm bg-white">
                                <option value="">Select condition...</option>
                                <option value="N">New</option>
                                <option value="U">Used</option>
                            </select>
                        </div>
                        <div>
                            <label class="block mb-2 font-medium text-gray-700">Notes:</label>
                            <textarea id="notes" name="notes" rows="3" placeholder="Any notes about this minifig..." class="w-full p-3 border border-gray-300 rounded-md text-sm resize-vertical"></textarea>
                        </div>
                    </div>
                    
                    <div class="flex justify-center gap-4 mt-8">
                        <button type="button" onclick="closeMinifigDetailsModal()" class="bg-gray-500 hover:bg-gray-600 text-white px-6 py-3 rounded-lg font-semibold transition-colors">
                            Cancel
                        </button>
                        <button type="button" onclick="submitMinifigDetails(event)" class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg font-semibold transition-colors">
                            Add Minifig
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    const container = document.getElementById('modal-container') || document.body;
    container.innerHTML = modalHtml;
    
    // Load parts for the modal
    populateModalPartsList();
}

// Modal parts population is now handled by server-side templates

/**
 * Load part image
 */
function loadPartImage(partId, itemType, partNo, colorId) {
    fetch('/partial-minifigs-lists/part-picture', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'item_type=' + encodeURIComponent(itemType) + 
              '&item_id=' + encodeURIComponent(partNo) + 
              '&color_id=' + encodeURIComponent(colorId)
    })
    .then(response => response.json())
    .then(data => {
        const imageContainer = document.getElementById(partId + '-image');
        
        if (data && data.meta && data.meta.code === 200 && data.data) {
            let imageData = Array.isArray(data.data) ? data.data : [data.data];
            
            if (imageData.length > 0) {
                const firstImage = imageData[0];
                const imageUrl = firstImage.thumbnail_url || firstImage.url;
                
                if (imageUrl && imageContainer) {
                    imageContainer.innerHTML = '<img src="' + imageUrl + '" alt="Part ' + partNo + '" class="max-w-full max-h-full object-contain rounded">';
                    imageContainer.classList.add('flex', 'items-center', 'justify-center');
                }
            }
        }
    })
    .catch(error => {
        console.error('Error loading part image:', error);
    });
}

/**
 * Load part pricing for modal parts
 */
function loadPartPricing(partId, itemType, partNo, colorId, quantity, itemId) {
    Promise.all([
        loadPartPricingByCondition(partId, itemType, partNo, colorId, 'N'),
        loadPartPricingByCondition(partId, itemType, partNo, colorId, 'U')
    ])
    .then(([newPrice, usedPrice]) => {
        updatePriceDisplay(itemId, partId + '-price-new', newPrice);
        updatePriceDisplay(itemId, partId + '-price-used', usedPrice);
        updatePartOutValues(itemId, newPrice, usedPrice, parseInt(quantity));
    })
    .catch(error => {
        console.error('Error loading part pricing:', error);
    });
}

/**
 * Populate modal parts list via API (fallback when no parts provided)
 */
function populateModalPartsList() {
    const minifigData = window.currentMinifigData;
    if (!minifigData || !minifigData.itemId) {
        console.error('No minifig data available for modal');
        return;
    }
    
    // Check if parts are already rendered
    const existingParts = document.querySelectorAll('.modal-part-card');
    if (existingParts.length > 0) {
        console.log('Parts already rendered in modal, skipping API call');
        return;
    }
    
    const itemId = minifigData.itemId;
    
    fetch('/partial-minifigs-lists/minifig-parts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'bricklink_id=' + encodeURIComponent(itemId)
    })
    .then(response => response.json())
    .then(data => {
        if (data && data.meta && data.meta.code === 200 && data.data && data.data.length > 0) {
            const processedParts = processPartsData(data.data);
            // Parts are now handled by server-side templates
        } else {
            const partsListElement = document.getElementById('modal-parts-list');
            const partsCountElement = document.getElementById('modal-parts-count');
            partsListElement.innerHTML = '<div class="text-gray-400 text-center">No parts data available</div>';
            partsCountElement.textContent = '0';
        }
    })
    .catch(error => {
        console.error('Error loading modal parts:', error);
        document.getElementById('modal-parts-list').innerHTML = '<div class="text-red-500 text-center">Error loading parts</div>';
        document.getElementById('modal-parts-count').textContent = '0';
    });
}

/**
 * Close minifig details modal
 */
function closeMinifigDetailsModal() {
    const overlay = document.getElementById('minifig-details-overlay');
    if (overlay) {
        overlay.remove();
    }
}

/**
 * Toggle all modal parts selection
 */
function toggleAllModalParts() {
    const selectAllButton = document.getElementById('modal-select-all');
    const partCards = document.querySelectorAll('.modal-part-card');
    
    const selectedCards = document.querySelectorAll('.modal-part-card.selected');
    const allSelected = selectedCards.length === partCards.length;
    
    partCards.forEach(card => {
        const indicator = card.querySelector('.modal-selection-indicator');
        const checkmark = card.querySelector('.modal-checkmark');
        
        if (allSelected) {
            // Deselect all
            card.classList.remove('selected', 'border-blue-500', 'bg-blue-50', 'shadow-lg');
            card.classList.add('border-gray-200', 'bg-white');
            if (indicator) {
                indicator.classList.remove('bg-blue-500', 'border-blue-500');
                indicator.classList.add('bg-white', 'border-gray-400');
            }
            if (checkmark) {
                checkmark.classList.add('hidden');
            }
        } else {
            // Select all
            card.classList.add('selected', 'border-blue-500', 'bg-blue-50', 'shadow-lg');
            card.classList.remove('border-gray-200', 'bg-white');
            if (indicator) {
                indicator.classList.remove('bg-white', 'border-gray-400');
                indicator.classList.add('bg-blue-500', 'border-blue-500');
            }
            if (checkmark) {
                checkmark.classList.remove('hidden');
            }
        }
    });
    
    updateModalSelectAllButtonText();
}

/**
 * Toggle individual modal part card selection
 */
function toggleModalPartCard(partId) {
    const card = document.getElementById(partId + '-card');
    if (!card) return;
    
    const indicator = card.querySelector('.modal-selection-indicator');
    const checkmark = card.querySelector('.modal-checkmark');
    
    if (card.classList.contains('selected')) {
        // Deselect
        card.classList.remove('selected', 'border-blue-500', 'bg-blue-50', 'shadow-lg');
        card.classList.add('border-gray-200', 'bg-white');
        if (indicator) {
            indicator.classList.remove('bg-blue-500', 'border-blue-500');
            indicator.classList.add('bg-white', 'border-gray-400');
        }
        if (checkmark) {
            checkmark.classList.add('hidden');
        }
    } else {
        // Select
        card.classList.add('selected', 'border-blue-500', 'bg-blue-50', 'shadow-lg');
        card.classList.remove('border-gray-200', 'bg-white');
        if (indicator) {
            indicator.classList.remove('bg-white', 'border-gray-400');
            indicator.classList.add('bg-blue-500', 'border-blue-500');
        }
        if (checkmark) {
            checkmark.classList.remove('hidden');
        }
    }
    
    updateModalSelectAllButtonText();
}

/**
 * Update modal select all button text
 */
function updateModalSelectAllButtonText() {
    const selectAllButton = document.getElementById('modal-select-all');
    const partCards = document.querySelectorAll('.modal-part-card');
    const selectedCards = document.querySelectorAll('.modal-part-card.selected');
    
    if (!selectAllButton || !partCards) return;
    
    if (selectedCards.length === 0) {
        selectAllButton.textContent = 'Select All';
    } else if (selectedCards.length === partCards.length) {
        selectAllButton.textContent = 'Deselect All';
    } else {
        selectAllButton.textContent = 'Select All';
    }
}

/**
 * Submit minifig details form
 */
function submitMinifigDetails(event) {
    event.preventDefault();
    
    const form = event.target;
    const formData = new FormData(form);
    const minifigData = window.currentMinifigData;
    
    if (!minifigData || !minifigData.itemId) {
        alert('Missing minifig data. Please try again.');
        return false;
    }
    
    // Collect selected parts
    const selectedCards = document.querySelectorAll('.modal-part-card.selected');
    const selectedParts = Array.from(selectedCards).map(card => ({
        PartNo: card.dataset.partNo,
        PartName: card.dataset.partName,
        Quantity: parseInt(card.dataset.quantity),
        ColorId: parseInt(card.dataset.colorId),
        ItemType: card.dataset.itemType,
        condition: formData.get('condition') || ''
    }));
    
    const referenceId = formData.get('reference_id') || '';
    const condition = formData.get('condition') || '';
    const notes = formData.get('notes') || '';
    
    console.log('Submitting minifig with details:', {
        itemId: minifigData.itemId,
        referenceId: referenceId,
        condition: condition,
        notes: notes,
        selectedParts: selectedParts
    });
    
    closeMinifigDetailsModal();
    
    // Disable add button while processing
    const addButton = document.getElementById('add-minifig-' + minifigData.itemId);
    if (addButton) {
        addButton.disabled = true;
        addButton.textContent = 'Adding...';
    }
    
    // Submit to server
    fetch('/partial-minifigs-lists/add-minifig-with-parts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: 'list_id=' + encodeURIComponent(minifigData.listId) +
              '&minifig_id=' + encodeURIComponent(minifigData.itemId) +
              '&reference_id=' + encodeURIComponent(referenceId) +
              '&condition=' + encodeURIComponent(condition) +
              '&notes=' + encodeURIComponent(notes) +
              '&selected_parts=' + encodeURIComponent(JSON.stringify(selectedParts))
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.text();
    })
    .then(html => {
        console.log('Successfully added minifig and parts');
        
        // Update the list
        const listItemsSection = document.querySelector('[data-section="list-items"]');
        if (listItemsSection) {
            listItemsSection.innerHTML = html;
        }
        
        // Show success message
        if (typeof showNotification === 'function') {
            showNotification('Successfully added ' + minifigData.itemId + ' with ' + selectedParts.length + ' parts!', 'success');
        } else {
            alert('Successfully added ' + minifigData.itemId + ' with ' + selectedParts.length + ' parts!');
        }
    })
    .catch(error => {
        console.error('Error adding minifig and parts:', error);
        alert('Failed to add minifig and parts: ' + error.message);
    })
    .finally(() => {
        // Re-enable button
        if (addButton) {
            addButton.disabled = false;
            addButton.textContent = 'Add';
        }
    });
    
    return false;
}

// Utility functions
function showNotification(message, type) {
    if (typeof window.showNotification === 'function') {
        window.showNotification(message, type);
    } else {
        alert(message);
    }
}

function showErrorMessage(itemId, message) {
    const errorElement = document.getElementById('error-message-' + itemId);
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.style.display = 'block';
    }
}

function hideErrorMessage(itemId) {
    const errorElement = document.getElementById('error-message-' + itemId);
    if (errorElement) {
        errorElement.style.display = 'none';
    }
}