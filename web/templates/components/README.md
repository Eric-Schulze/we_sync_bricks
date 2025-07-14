# Template Components

This directory contains reusable template components organized by logical functionality.

## Structure

```
components/
├── cards/
│   ├── bricklink_item_card.html         # Complete Bricklink item with parts selection
│   ├── minifig_picture.html             # Minifig image with loading state  
│   ├── minifig_info.html                # Minifig name, ID, and pricing display
│   └── price_display.html               # New/Used price grid
├── forms/
│   ├── add_minifig_button.html          # Action button with validation
│   ├── bricklink_search_form.html       # Search input and results container
│   └── minifig_details_form.html        # Form for reference ID, condition, notes
├── modals/
│   ├── base_modal.html                  # Reusable modal wrapper
│   ├── minifig_details_modal.html       # Modal for parts selection and details
│   └── new_minifig_modal.html           # Search modal for adding minifigs
├── parts/
│   ├── bricklink_item_scripts.html      # JavaScript includes
│   ├── minifig_part.html                # Individual part card for modal
│   └── parts_list_container.html        # Collapsible parts list with toggle
└── helpers.html                         # Template helper functions
```

## Component-Based Architecture

**Individual components that work together, avoiding duplication:**

### Core Components

```html
{{/* Minifig picture with loading state */}}
{{template "minifig-picture" (dict "ItemNo" .Item.No)}}

{{/* Minifig info with name, ID, button */}}
{{template "minifig-info" (dict "Name" .Item.Name "ItemNo" .Item.No)}}

{{/* Price display grid */}}
{{template "price-display" (dict "ItemNo" .Item.No)}}

{{/* Add button with validation */}}
{{template "add-minifig-button" (dict "ItemNo" .Item.No)}}
```

### Complex Components

```html
{{/* Complete Bricklink item card */}}
{{template "bricklink-item-card" .}}

{{/* Modal with parts selection and form */}}
{{template "minifig-details-modal" .}}

{{/* Collapsible parts list */}}
{{template "parts-list-container" (dict "ItemNo" .Item.No)}}
```

## Legacy Components

### Still Active

```html
{{/* Complete Bricklink item card - now uses consolidated template */}}
{{template "bricklink-item-card" .}}

{{/* Search modal */}}
{{template "new-minifig-modal" .}}

{{/* Search form */}}
{{template "bricklink-search-form" .}}
```

### Deprecated Components

These templates have been removed and consolidated:
- ~~`minifig-picture`~~ - Now part of consolidated template
- ~~`minifig-info`~~ - Now part of consolidated template  
- ~~`price-display`~~ - Now part of consolidated template
- ~~`minifig-part`~~ - Now part of consolidated template
- ~~`add-minifig-button`~~ - Now part of consolidated template
- ~~`minifig-details-form`~~ - Now part of consolidated template
- ~~`minifig-details-modal`~~ - Now part of consolidated template
- ~~`parts-list-container`~~ - Now part of consolidated template
- ~~`bricklink-item-scripts`~~ - Replaced by `/static/js/minifig-parts.js`

## Features

- **Responsive Design**: All components adapt to container width
- **Template-Based**: Uses `<template>` tags for reusable DOM elements
- **Data-Driven**: JavaScript builds data objects instead of HTML strings
- **Interactive Elements**: Card selection, toggleable parts lists, modal overlays
- **Error Handling**: Built-in validation and user feedback
- **Accessibility**: Proper focus management and keyboard navigation
- **Performance**: No string concatenation for HTML generation

## JavaScript Integration

The `/static/js/minifig-parts.js` file provides:

- **Template Utilization**: Uses `cloneNode()` to populate templates
- **Data Management**: Clean separation between data and presentation
- **Auto-loading**: Pictures, pricing, and parts data
- **Part Selection**: Click-to-select card interface
- **Validation**: Error messages for empty selections
- **API Integration**: Fetch calls to backend services