# Template Components

This directory contains reusable template components organized by logical functionality.

## Structure

```
components/
├── cards/
│   ├── bricklink_item_card.html    # Complete Bricklink item with parts selection
│   ├── minifig_info.html           # Minifig name, ID, and pricing display
│   ├── minifig_picture.html        # Minifig image container
│   └── price_display.html          # New/Used price grid
├── forms/
│   ├── add_minifig_button.html     # Action button with error handling
│   └── bricklink_search_form.html  # Search input and results container
├── modals/
│   ├── base_modal.html             # Reusable modal wrapper
│   └── new_minifig_modal.html      # Specific modal for adding minifigs
├── parts/
│   ├── bricklink_item_scripts.html # JavaScript functionality for Bricklink items
│   └── parts_list_container.html   # Collapsible parts list with toggle
└── helpers.html                    # Template helper functions
```

## Usage

### Basic Components

```html
{{/* Minifig picture with loading state */}}
{{template "minifig-picture" (dict "ItemNo" .Item.No)}}

{{/* Price display grid */}}
{{template "price-display" (dict "ItemNo" .Item.No)}}

{{/* Add button with validation */}}
{{template "add-minifig-button" (dict "ItemNo" .Item.No "ButtonText" "Custom Text")}}
```

### Complex Components

```html
{{/* Complete Bricklink item card */}}
{{template "bricklink-item-card" .}}

{{/* Modal with search form */}}
{{template "new-minifig-modal" .}}
```

### Data Passing

Components use the `dict` template function to pass named parameters:

```html
{{template "component-name" (dict "Key1" .Value1 "Key2" "literal value")}}
```

## Features

- **Responsive Design**: All components adapt to container width
- **Interactive Elements**: Card selection, toggleable parts lists, modal overlays
- **Error Handling**: Built-in validation and user feedback
- **Accessibility**: Proper focus management and keyboard navigation
- **Reusability**: Components can be combined and customized via parameters

## Customization

Most components support customization through template variables:

- `MaxWidth`: Modal maximum width (default: "600px")
- `ButtonText`: Custom button text
- `ShowCloseButton`: Whether to show close button (default: true)
- `CloseButtonText`: Custom close button text (default: "Close")

## JavaScript Integration

The `bricklink_item_scripts.html` component provides:

- **Auto-loading**: Pictures, pricing, and parts data
- **Part Selection**: Click-to-select card interface
- **Validation**: Error messages for empty selections
- **API Integration**: Fetch calls to backend services