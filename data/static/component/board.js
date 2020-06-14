'use strict';
import { h } from '/js/preact.js';

function Board(props){
	let children = props.children || []
	if (!Array.isArray(children)){
		children = [children]
	}
	let body
	if (props.style=="list"){
		body = children.map(c=>h('div',{className:'board-list-element'},c))
	}else{
		body = [h('div',{className:'board-list-element'},...children)]
	}
	return h('div',{className:'content-board'},
			h('div',{className:'board-title'},
					h('div',null,props.title),
					props.titleIcon && h('img',{src:props.titleIcon,onClick:props.iconOnClick}),
					props.icons && h('div',null,...props.icons)
			),
			h('div',{className:'board-body'},...body)
	)
}

export {Board}