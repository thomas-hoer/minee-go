'use strict'
import { h } from '/js/preact.js'

const loadComponent = async function(pathname, options){
	const comp = await import(pathname+"component.js")
	return h(comp.Component,{...options, path:pathname})
}
const loadComponents = async function(pathname, options, filter){
	const elements = await fetch(pathname+"?json").then(r=>r.json())
	let result = []
	let filterOK
	let noFilter = true
	if (filter){
		noFilter = false
		filterOK = async (path)=>{
			const type = await fetch(path+"type").then(res=>res.text())
			return type == filter
		}
	}
	for(const element of elements){
		if(element.endsWith("/") && ( noFilter||await filterOK(pathname+element))){
			result.push(await loadComponent(pathname+element, options))
		}
	}
	return result
}

export {loadComponent,loadComponents}