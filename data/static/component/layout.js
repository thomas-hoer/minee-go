'use strict'
import { h, Fragment } from '/js/preact.js'
import { Menu } from '/component/menu.js'
import { Text } from '/component/text.js'

function Layout(props){

	return h(Fragment,null,
			h('div',{className:'header'},
					h('div',null,props.title),
			),
			h('div',{className:'body'},
					h('div',{className:'menu'},h(Menu)),
					h('div',{className:'content-wrapper'},props.children)
			),
			h('div',{className:'footer'})
	)
}

export {Layout}