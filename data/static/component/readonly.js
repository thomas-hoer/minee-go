'use strict';
import { h } from '/js/preact.js';

function Readonly(props){
	return h('div',{className:'input-readonly'},props.property.get())
}

export {Readonly}