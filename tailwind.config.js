/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/*.{html, js}"],
  theme: {
        extend: {
            colors: {
            'text': '#e8edf7',
            'background': '#060c14',
            'primary': '#3b4f72',
            'secondary': '#214687',
            'accent': '#29497f',
            },
        },
  },
  plugins: [],
}

