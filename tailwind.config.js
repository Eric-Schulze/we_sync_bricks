/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.html",
    "./web/static/js/**/*.js",
    "./partial_minifigs/**/*.go",
    "./internal/**/*.go"
  ],
  theme: {
    extend: {
      colors: {
        'lego-red': '#d40000',
        'lego-yellow': '#ffc70a',
        'lego-blue': '#0066cc',
        'lego-green': '#00852b',
        'new-price': {
          50: '#eff6ff',
          500: '#3b82f6',
          700: '#1d4ed8',
        },
        'used-price': {
          50: '#f0fdf4',
          500: '#22c55e',
          700: '#15803d',
        },
        'part-out-value': {
          50: '#faf5ff',
          500: '#a855f7',
          700: '#7c3aed',
        },
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}