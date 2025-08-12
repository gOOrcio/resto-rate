<script lang="ts">
    import clients from '$lib/client/client';
    import type { Place, Suggestion } from '$lib/client/generated/google_maps/v1/google_maps_service_pb';
    import { onMount, onDestroy } from 'svelte';
    import InputSv from './InputSv.svelte';
    import { v4 as uuidv4 } from 'uuid';

    function randomUUID(): string {
        return uuidv4();
    }

    let autocompleteSessionToken = randomUUID();
    let input = $state('');
    let suggestions = $state<Suggestion[]>([]);
    let isLoading = $state(false);
    let selectedIndex = $state(-1);
    let showSuggestions = $state(false);
    let queryPrediction = $state('');
    let debounceTimer: ReturnType<typeof setTimeout> | null = null;
    
    let { onPlaceSelected } = $props<{
        onPlaceSelected?: (place: Place) => void;
    }>();

    function debouncedAutocomplete(input: string) {
        if (debounceTimer) {
            clearTimeout(debounceTimer);
        }
        
        if (input.length < 2) {
            suggestions = [];
            showSuggestions = false;
            queryPrediction = '';
            autocompleteSessionToken = randomUUID();
            return;
        }
        
        if (input.length >= 2 && !isLoading) {
            isLoading = true;
        }
        
        debounceTimer = setTimeout(() => {
            if (input.length >= 2) {
                performAutocomplete(input);
            }
        }, 300);
    }

    async function performAutocomplete(input: string) {
        if (input.length < 2) return;
        
        isLoading = true;
        try {
            const response = await clients.googleMaps.autocompletePlaces({
                input,
                languageCode: 'pl',
                includedRegionCodes: ['pl'],
                sessionToken: autocompleteSessionToken,
                includeQueryPrediction: true,
            });
            
            suggestions = response.suggestions || [];
            showSuggestions = suggestions.length > 0;
            
            const querySuggestion = suggestions.find(s => s.queryPrediction);
            if (querySuggestion?.queryPrediction?.text?.text) {
                queryPrediction = querySuggestion.queryPrediction.text.text;
            } else {
                queryPrediction = '';
            }
            
            selectedIndex = -1;
        } catch (error) {
            console.error('Autocomplete error:', error);
            suggestions = [];
            showSuggestions = false;
        } finally {
            isLoading = false;
        }
    }

    async function getPlaceDetails(name: string) {
        try {
            const response = await clients.googleMaps.getRestaurantDetails({
                name: name,
                languageCode: 'pl',
                regionCode: 'pl',
                sessionToken: autocompleteSessionToken,
            });
            
            if (onPlaceSelected) {
                onPlaceSelected(response);
            }
            
            suggestions = [];
            showSuggestions = false;
            queryPrediction = '';
            input = response.displayName?.text || response.name || '';
            
            autocompleteSessionToken = randomUUID();
            
        } catch (error) {
            console.error('Get place details error:', error);
        }
    }

    function handleInputChange(event: Event) {
        const target = event.target as HTMLInputElement;
        const newValue = target.value;
        
        input = newValue;
        
        if (!newValue.trim()) {
            suggestions = [];
            showSuggestions = false;
            queryPrediction = '';
            isLoading = false;
            return;
        }
        
        debouncedAutocomplete(newValue);
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (!showSuggestions) return;
        
        switch (event.key) {
            case 'ArrowDown':
                event.preventDefault();
                selectedIndex = Math.min(selectedIndex + 1, suggestions.length - 1);
                break;
            case 'ArrowUp':
                event.preventDefault();
                selectedIndex = Math.max(selectedIndex - 1, -1);
                break;
            case 'Enter':
                event.preventDefault();
                if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
                    selectSuggestion(suggestions[selectedIndex]);
                }
                break;
            case 'Escape':
                showSuggestions = false;
                selectedIndex = -1;
                break;
        }
    }

    function selectSuggestion(suggestion: Suggestion) {
        if (suggestion.placePrediction?.place) {
            getPlaceDetails(suggestion.placePrediction.place);
        }
    }

    function getSuggestionText(suggestion: Suggestion): string {
        if (suggestion.placePrediction?.structuredFormat?.mainText?.text) {
            return suggestion.placePrediction.structuredFormat.mainText.text;
        }
        if (suggestion.placePrediction?.text?.text) {
            return suggestion.placePrediction.text.text;
        }
        return '';
    }

    function getSuggestionSubtext(suggestion: Suggestion): string {
        if (suggestion.placePrediction?.structuredFormat?.secondaryText?.text) {
            return suggestion.placePrediction.structuredFormat.secondaryText.text;
        }
        return '';
    }

    onMount(() => {
        autocompleteSessionToken = randomUUID();
    });

    onDestroy(() => {
        if (debounceTimer) {
            clearTimeout(debounceTimer);
            debounceTimer = null;
        }
    });
</script>

        <div class="relative w-full max-w-md">
        <div class="relative flex items-center">
            <InputSv
                type="text" 
                bind:value={input} 
                oninput={handleInputChange}
                onkeydown={handleKeyDown}
                placeholder="Search for restaurants..."
                autocomplete="off"
                class="
                    w-full pr-10
                    bg-no-repeat
                    bg-[length:60px_60px]
                    bg-[url('/GoogleMaps_Logo_Gray.svg')]
                    bg-[position:calc(100%-2.25rem)_50%]
                    "
            />
            {#if isLoading}
                <div class="absolute right-3 flex items-center">
                    <div class="animate-spin rounded-full h-4 w-4 border-2 border-gray-300 border-t-primary-500"></div>
                </div>
            {/if}
        </div>
        
        {#if showSuggestions && suggestions.length > 0}
            <div class="absolute top-full left-0 right-0 bg-white border-2 border-gray-200 border-t-0 rounded-b-lg shadow-lg max-h-80 overflow-y-auto z-50">
                {#each suggestions as suggestion, index}
                    {#if suggestion.placePrediction}
                        <div 
                            class="p-3 cursor-pointer border-b border-gray-100 last:border-b-0 transition-colors duration-200 hover:bg-gray-50 {index === selectedIndex ? 'bg-gray-50' : ''}"
                            onclick={() => selectSuggestion(suggestion)}
                            onkeydown={(e) => e.key === 'Enter' && selectSuggestion(suggestion)}
                            onmouseenter={() => selectedIndex = index}
                            tabindex="0"
                            role="button"
                            aria-label="Select {getSuggestionText(suggestion)}"
                        >
                            <div class="font-medium text-gray-900 mb-1">
                                {getSuggestionText(suggestion)}
                            </div>
                            {#if getSuggestionSubtext(suggestion)}
                                <div class="text-sm text-gray-500">
                                    {getSuggestionSubtext(suggestion)}
                                </div>
                            {/if}
                        </div>
                    {/if}
                {/each}
            </div>
        {/if}
        
        {#if queryPrediction && input.length > 0}
            <div class="absolute top-1/2 left-4 transform -translate-y-1/2 pointer-events-none text-base text-gray-500 z-10 {showSuggestions ? 'hidden' : ''}">
                <span class="text-transparent">{input}</span>
                <span class="text-gray-500 opacity-60">{queryPrediction.substring(input.length)}</span>
            </div>
        {/if}
        
        {#if input.length > 0 && input.length < 2}
            <div class="absolute top-full left-0 right-0 mt-1 p-2 text-sm text-gray-500 bg-gray-50 border border-gray-200 rounded">
                Type at least 2 characters to search...
            </div>
        {/if}
    </div>

