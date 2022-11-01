const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./index.html', './src/**/*.{vue,js,ts}'],
    theme: {
        extend: {
            fontFamily: {
                sans: ['"Inter var"', ...defaultTheme.fontFamily.sans],
            },
            width: {
                'fit': 'fit-content',
                '68': '17rem'
            },
            height: {
                'fit': 'fit-content'
            },
            fontSize: {
                'huge': '10rem'
            },
            minHeight: {
                '9/10': '90%',
                '4/5': '80%',
                '7/10': '70%',
                '1/2': '50%',
                '1/3': '33.333%',
                '44': '11rem',
                '32': '8rem',
                '28': '7rem',
                '20': '5rem',
                '18': '4.5rem',
                '16': '4rem',
            },
        },
    },
    plugins: [
        require('@tailwindcss/forms'),
        require('@tailwindcss/typography'),
        require('@tailwindcss/line-clamp'),
        require('@tailwindcss/aspect-ratio'),
    ],
}
