/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.html",
    "./web/static/js/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        'lego-red': '#d40000',
        'lego-yellow': '#ffc70a',
        'lego-blue': '#0066cc',
        'lego-green': '#00852b',
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}