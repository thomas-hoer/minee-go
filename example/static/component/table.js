'use strict'
import { h } from '/js/preact.js'

function Table(props){
	let children = props.children || []
	let body = children.map(c=>h('div',null,c))
	return h('div',{
		style:{
			...props.style, 
			display: "grid", 
			gridTemplateColumns: props.gridTemplate
		}
	},body)
}

export {Table}