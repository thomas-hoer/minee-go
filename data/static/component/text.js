'use strict';
import { h } from '/js/preact.js';

function Text(props){
	const type = props.type || 'text'
	return h('input',{
		value:props.property.get(),
		onChange:(ev)=>props.property.set(event.target.value),
		type:type,
		placeholder:props.placeholder,
		})
}

export {Text}