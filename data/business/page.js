'use strict'
import { h } from '/js/preact.js'
import { Layout } from '/component/layout.js'

function Page(props){
	const layoutOptions = {
			title:"MinEE-GO",
	}
	return h(Layout,layoutOptions,"Hello World!")
}
export {Page}