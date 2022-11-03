// @ts-ignore
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
// @ts-ignore
import { library } from '@fortawesome/fontawesome-svg-core'
// @ts-ignore
import { fas } from '@fortawesome/free-solid-svg-icons'

library.add(fas)

export default (app: any) => {
    app.component('font-awesome-icon', FontAwesomeIcon)
}