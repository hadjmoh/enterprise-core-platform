/**
 * Cryptographically secure random number utilities.
 */

/**
 * Returns a secure random float between 0 (inclusive) and 1 (exclusive).
 * Equivalent to Math.random() but cryptographically secure.
 */
export const secureRandom = (): number => {
    const array = new Uint32Array(1);
    window.crypto.getRandomValues(array);
    return array[0] / (0xffffffff + 1);
};

/**
 * Returns a secure random integer between min (inclusive) and max (inclusive).
 */
export const secureRandomInt = (min: number, max: number): number => {
    const range = max - min + 1;
    const bytesNeeded = Math.ceil(Math.log2(range) / 8);
    const maxVal = Math.pow(256, bytesNeeded);
    const array = new Uint8Array(bytesNeeded);

    let value;
    do {
        window.crypto.getRandomValues(array);
        value = 0;
        for (let i = 0; i < bytesNeeded; i++) {
            value = (value << 8) + array[i];
        }
    } while (value >= maxVal - (maxVal % range)); // Avoid modulo bias

    return min + (value % range);
};
