'use strict';
import { h } from '/js/preact.js';
import { Readonly } from './readonly.js'

function DatePicker(props){
	if (props.readOnly){
		const val = (new Date(props.property.get())).toLocaleDateString()
		return h(Readonly,{property:{get:()=>val}})
	}
	let val = props.property.get()
	const input = h('input',{
		value:val&&val.substring(0,10),
		onChange:(ev)=>props.property.set(event.target.value),
		type:"date"
	})
	
	return props.label?h('label',null,h('div',null,props.label),input):input

}

export {DatePicker}