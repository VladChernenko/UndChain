'''
This script is meant to read the chain rules file for the validators
and partners.
'''

from typing import Tuple
from console_colors import Console_Colors as cli

def known_validator(account_info: Tuple[str, str], chain: Tuple[int, int]) -> bool:
    '''
    This function returns true if the validator is a known validator for the 
    chain that has been specified, else false
    '''

    print(f'Looking at chain ID: {chain} to see if I am a known validator. My account info is: {account_info}')
    print(f'{cli.RED}[Known Validator]: need to be fully implemented! Returns True, if you are on the main chain on the test net: {cli.RESET}')
    if chain == (0, 0):
        return True
    else:
        return False
