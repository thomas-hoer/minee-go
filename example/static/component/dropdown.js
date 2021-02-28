'use strict'
import { h } from '/js/preact.js'
import { Readonly } from './readonly.js'

function Dropdown(props){
	if (props.readOnly){
		const propGet = props.property.get()
		const val = props.options.find(e=>e.value==propGet)
		const property={get:()=>val&&val.label}
		return h(Readonly,{property:property})
	}
	return h('select',{value:props.property.get(),onChange:ev=>props.property.set(ev.target.value)},props.options.map(o=>
	h('option',{value:o.value},o.label)
	))
}

export {Dropdown}