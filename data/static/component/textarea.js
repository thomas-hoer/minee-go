'use strict';
import { h } from '/js/preact.js';

function Textarea(props){
	return h('textarea',{
		value:props.property.get(),
		onInput:(ev)=>props.property.set(event.target.value),
		placeholder:props.placeholder,
		})
}

export {Textarea}